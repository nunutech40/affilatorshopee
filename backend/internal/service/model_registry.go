package service

type ModelInfo struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Name     string `json:"name"`
	Free     bool   `json:"free"`
	Context  int    `json:"context_window"`
	Note     string `json:"note"`
}

func ListModels() []ModelInfo {
	return []ModelInfo{
		{ID: "codex/gpt-5.6-luna", Provider: "codex", Name: "GPT-5.6 Luna (Codex CLI)", Free: false, Context: 400000, Note: "Local Codex CLI bridge"},
		{ID: "codex/gpt-5.6-sol", Provider: "codex", Name: "GPT-5.6 Sol (Codex CLI)", Free: false, Context: 400000, Note: "Local Codex CLI bridge"},
		{ID: "stealth/ox-alpha", Provider: "openrouter", Name: "Ox Alpha (OpenRouter)", Free: false, Context: 200000, Note: "OpenRouter"},
		{ID: "openrouter/minimax/minimax-m3:free", Provider: "openrouter", Name: "MiniMax M3 Free", Free: true, Context: 196608, Note: "OpenRouter"},
		{ID: "opencode/muse-spark-1.2-contributor-free", Provider: "opencode", Name: "Muse Spark 1.2 Contributor Free", Free: true, Context: 1048576, Note: "OpenCode"},
		{ID: "opencode/muse-spark-1.2", Provider: "opencode", Name: "Muse Spark 1.2", Free: false, Context: 1048576, Note: "OpenCode"},
		{ID: "opencode/gpt-5.6-luna", Provider: "opencode", Name: "GPT-5.6 Luna", Free: false, Context: 1050000, Note: "OpenCode"},
		{ID: "opencode/deepseek-v4-flash-free", Provider: "opencode", Name: "DeepSeek V4 Flash Free", Free: true, Context: 200000, Note: "OpenCode"},
		{ID: "ag/gemini-3-flash", Provider: "9router", Name: "Gemini 3 Flash", Free: false, Context: 1048576, Note: "9router"},
		{ID: "ag/gemini-3-flash-agent", Provider: "9router", Name: "Gemini 3 Flash Agent", Free: false, Context: 1048576, Note: "9router"},
		{ID: "ag/gemini-3.5-flash-low", Provider: "9router", Name: "Gemini 3.5 Flash Low", Free: false, Context: 1048576, Note: "9router"},
		{ID: "ag/gemini-3.5-flash-extra-low", Provider: "9router", Name: "Gemini 3.5 Flash Extra Low", Free: false, Context: 1048576, Note: "9router"},
		{ID: "ds/deepseek-v4-flash", Provider: "9router", Name: "DeepSeek V4 Flash", Free: false, Context: 200000, Note: "9router"},
		{ID: "ocg/glm-5.2", Provider: "9router", Name: "GLM 5.2", Free: false, Context: 200000, Note: "9router"},
	}
}
