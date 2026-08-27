package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type executeRequest struct {
	Model        string `json:"model"`
	Instructions string `json:"instructions"`
	Input        string `json:"input"`
}

type chatCompletionRequest struct {
	Model       string          `json:"model"`
	Messages    []chatMessage   `json:"messages"`
	Temperature json.RawMessage `json:"temperature"`
	MaxTokens   json.RawMessage `json:"max_tokens"`
	Stream      bool            `json:"stream"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIError struct {
	Error openAIErrorBody `json:"error"`
}

type openAIErrorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

var token = strings.TrimSpace(os.Getenv("CODEX_BRIDGE_TOKEN"))

func main() {
	if token == "" {
		log.Fatal("CODEX_BRIDGE_TOKEN wajib diisi")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("/v1/execute", execute)
	mux.HandleFunc("/v1/chat/completions", chatCompletions)
	addr := os.Getenv("CODEX_BRIDGE_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8787"
	}
	log.Printf("Codex bridge listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func authorize(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Authorization") != "Bearer "+token {
		writeOpenAIError(w, http.StatusUnauthorized, "unauthorized", "authentication_error")
		return false
	}
	return true
}

func execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Authorization") != "Bearer "+token {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var payload executeRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&payload); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	model := strings.TrimSpace(payload.Model)
	if model == "" || strings.ContainsAny(model, " \t\r\n") || strings.Contains(model, "/") {
		http.Error(w, "invalid Codex model", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Instructions) == "" || strings.TrimSpace(payload.Input) == "" {
		http.Error(w, "instructions and input are required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Minute)
	defer cancel()
	prompt := payload.Instructions + "\n\n" + payload.Input
	codexPath := strings.TrimSpace(os.Getenv("CODEX_CLI_PATH"))
	if codexPath == "" {
		codexPath, _ = exec.LookPath("codex")
	}
	if codexPath == "" {
		matches, _ := filepath.Glob(filepath.Join(os.Getenv("HOME"), ".vscode", "extensions", "openai.chatgpt-*", "bin", "macos-aarch64", "codex"))
		if len(matches) > 0 {
			codexPath = matches[len(matches)-1]
		}
	}
	if codexPath == "" {
		http.Error(w, "Codex CLI tidak tersedia; set CODEX_CLI_PATH", http.StatusBadGateway)
		return
	}
	cmd := exec.CommandContext(ctx, codexPath, "exec", "--ephemeral", "--sandbox", "read-only", "--skip-git-repo-check", "--json", "--model", model, prompt)
	cmd.Dir = os.TempDir()
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			http.Error(w, "Codex CLI timeout", http.StatusGatewayTimeout)
			return
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			http.Error(w, fmt.Sprintf("Codex CLI gagal: %s", strings.TrimSpace(string(exitErr.Stderr))), http.StatusBadGateway)
			return
		}
		http.Error(w, "Codex CLI tidak tersedia", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/x-ndjson")
	_, _ = w.Write(out)
}

func chatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeOpenAIError(w, http.StatusMethodNotAllowed, "method not allowed", "method_not_allowed")
		return
	}
	if !authorize(w, r) {
		return
	}

	var payload chatCompletionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20)).Decode(&payload); err != nil {
		writeOpenAIError(w, http.StatusBadRequest, "invalid request: "+err.Error(), "invalid_request_error")
		return
	}
	model := normalizeModel(payload.Model)
	if model == "" || strings.ContainsAny(model, " \t\r\n") || strings.Contains(model, "/") {
		writeOpenAIError(w, http.StatusBadRequest, "invalid Codex model", "invalid_request_error")
		return
	}
	if len(payload.Messages) == 0 {
		writeOpenAIError(w, http.StatusBadRequest, "messages are required", "invalid_request_error")
		return
	}
	prompt, err := messagesPrompt(payload.Messages)
	if err != nil {
		writeOpenAIError(w, http.StatusBadRequest, err.Error(), "invalid_request_error")
		return
	}

	content, status, err := runCodex(r.Context(), model, prompt)
	if err != nil {
		code := "codex_cli_error"
		if status == http.StatusGatewayTimeout {
			code = "codex_cli_timeout"
		}
		writeOpenAIError(w, status, "Codex CLI gagal: "+err.Error(), code)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":      "codex-" + uniqueID(),
		"object":  "chat.completion",
		"created": 0,
		"model":   model,
		"choices": []map[string]any{{
			"index": 0,
			"message": map[string]string{
				"role":    "assistant",
				"content": content,
			},
			"finish_reason": "stop",
		}},
	})
}

func normalizeModel(model string) string {
	return strings.TrimPrefix(strings.TrimSpace(model), "codex/")
}

func messagesPrompt(messages []chatMessage) (string, error) {
	var prompt strings.Builder
	for _, message := range messages {
		role := strings.TrimSpace(message.Role)
		switch role {
		case "system", "user", "assistant":
		default:
			return "", fmt.Errorf("unsupported message role: %s", role)
		}
		if prompt.Len() > 0 {
			prompt.WriteString("\n\n")
		}
		prompt.WriteString(role)
		prompt.WriteString(":\n")
		prompt.WriteString(message.Content)
	}
	return prompt.String(), nil
}

func runCodex(parent context.Context, model, prompt string) (string, int, error) {
	ctx, cancel := context.WithTimeout(parent, 3*time.Minute)
	defer cancel()
	codexPath := findCodex()
	if codexPath == "" {
		return "", http.StatusBadGateway, fmt.Errorf("Codex CLI tidak tersedia; set CODEX_CLI_PATH")
	}
	cmd := exec.CommandContext(ctx, codexPath, "exec", "--ephemeral", "--sandbox", "read-only", "--skip-git-repo-check", "--json", "--model", model, prompt)
	cmd.Dir = os.TempDir()
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() != nil {
			return "", http.StatusGatewayTimeout, fmt.Errorf("timeout")
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			if stderr == "" {
				stderr = err.Error()
			}
			return "", http.StatusBadGateway, errors.New(stderr)
		}
		return "", http.StatusBadGateway, err
	}
	content, err := parseAgentMessage(out)
	if err != nil {
		return "", http.StatusBadGateway, err
	}
	return content, http.StatusOK, nil
}

func findCodex() string {
	codexPath := strings.TrimSpace(os.Getenv("CODEX_CLI_PATH"))
	if codexPath == "" {
		codexPath, _ = exec.LookPath("codex")
	}
	if codexPath == "" {
		matches, _ := filepath.Glob(filepath.Join(os.Getenv("HOME"), ".vscode", "extensions", "openai.chatgpt-*", "bin", "macos-aarch64", "codex"))
		if len(matches) > 0 {
			codexPath = matches[len(matches)-1]
		}
	}
	return codexPath
}

func parseAgentMessage(body []byte) (string, error) {
	var final string
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	for {
		var event struct {
			Type string `json:"type"`
			Item struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"item"`
		}
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("parse output Codex CLI: %w", err)
		}
		if event.Type == "item.completed" && event.Item.Type == "agent_message" {
			final = event.Item.Text
		}
	}
	if strings.TrimSpace(final) == "" {
		return "", fmt.Errorf("Codex CLI response tidak memiliki agent message")
	}
	return final, nil
}

func writeOpenAIError(w http.ResponseWriter, status int, message, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(openAIError{Error: openAIErrorBody{Message: message, Type: "server_error", Code: code}})
}

func uniqueID() string {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err == nil {
		return fmt.Sprintf("%x", bytes)
	}
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
