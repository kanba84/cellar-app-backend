package llm

import "os"

const (
	defaultGeminiModel = "gemini-2.5-flash-lite"
	envGeminiModel     = "GEMINI_MODEL"
)

func geminiModelFromEnv() string {
	if model := os.Getenv(envGeminiModel); model != "" {
		return model
	}
	return defaultGeminiModel
}
