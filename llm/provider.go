package llm

import (
	"context"
	"fmt"
	"os"
)

// NewWineInfoProvider creates a WineInfoProvider from environment variables.
func NewWineInfoProvider() (WineInfoProvider, error) {
	providerType := os.Getenv("LLM_PROVIDER")
	if providerType == "" {
		providerType = "gemini"
	}

	switch providerType {
	case "gemini":
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
		}
		return NewGeminiProvider(context.Background(), apiKey, geminiModelFromEnv())
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", providerType)
	}
}
