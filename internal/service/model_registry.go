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
		{ID: "opencode/muse-spark-1.2-contributor-free", Provider: "opencode", Name: "Muse Spark 1.2 (Contributor Free)", Free: true, Context: 1048576, Note: "Gratis - data dipakai training Meta, default untuk reformat"},
		{ID: "opencode/muse-spark-1.2", Provider: "opencode", Name: "Muse Spark 1.2", Free: false, Context: 1048576, Note: "Berbayar $1.25/$4.25 per 1M"},
		{ID: "opencode/gpt-5.6-luna", Provider: "opencode", Name: "GPT-5.6 Luna", Free: false, Context: 1050000, Note: "OpenAI via Zen"},
		{ID: "opencode/deepseek-v4-flash-free", Provider: "opencode", Name: "DeepSeek V4 Flash (Free)", Free: true, Context: 200000, Note: "Gratis"},
		{ID: "opencode/glm-5-free", Provider: "opencode", Name: "GLM-5 (Free)", Free: true, Context: 204800, Note: "Gratis"},
		{ID: "opencode/glm-4.7-free", Provider: "opencode", Name: "GLM-4.7 (Free)", Free: true, Context: 204800, Note: "Gratis"},
		{ID: "opencode/kimi-k2.5-free", Provider: "opencode", Name: "Kimi K2.5 (Free)", Free: true, Context: 262144, Note: "Gratis"},
		{ID: "opencode/minimax-m2.1-free", Provider: "opencode", Name: "MiniMax M2.1 (Free)", Free: true, Context: 204800, Note: "Gratis"},
		{ID: "opencode/qwen3.6-plus-free", Provider: "opencode", Name: "Qwen3.6 Plus (Free)", Free: true, Context: 262144, Note: "Gratis"},
		{ID: "opencode/gemini-3-flash", Provider: "opencode", Name: "Gemini 3 Flash", Free: false, Context: 1048576, Note: "Google via Zen"},
		{ID: "opencode/grok-code", Provider: "opencode", Name: "Grok Code Fast 1 (Free)", Free: true, Context: 256000, Note: "Gratis"},
		{ID: "opencode/nemotron-3-nano-free", Provider: "opencode", Name: "Nemotron 3 Nano (Free)", Free: true, Context: 262144, Note: "Gratis"},
	}
}
