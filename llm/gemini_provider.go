package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements WineInfoProvider using the Gemini API.
type GeminiProvider struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiProvider creates a provider with a reused client and preconfigured model.
func NewGeminiProvider(ctx context.Context, apiKey, modelName string) (*GeminiProvider, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("gemini api key is empty")
	}
	if strings.TrimSpace(modelName) == "" {
		return nil, fmt.Errorf("gemini model name is empty")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	model := client.GenerativeModel(modelName)
	configureGenerativeModel(model)

	return &GeminiProvider{
		client: client,
		model:  model,
	}, nil
}

func configureGenerativeModel(model *genai.GenerativeModel) {
	model.SetTemperature(0.1)
	model.SetTopK(20)
	model.SetTopP(0.8)
	model.SetMaxOutputTokens(256)
	model.ResponseMIMEType = "application/json"
}

// Close releases the underlying Gemini client.
func (gp *GeminiProvider) Close() error {
	if gp == nil || gp.client == nil {
		return nil
	}
	return gp.client.Close()
}

// FetchWineInfo fetches wine information for the given lookup key.
func (gp *GeminiProvider) FetchWineInfo(ctx context.Context, key WineLookupKey) (*WineInfoResult, error) {
	if gp == nil {
		return nil, fmt.Errorf("gemini provider is nil")
	}
	if gp.model == nil {
		return nil, fmt.Errorf("gemini model is not configured")
	}

	wineName := strings.TrimSpace(key.Name)
	if wineName == "" {
		return nil, fmt.Errorf("wine name is empty")
	}

	return withRetry(ctx, func(callCtx context.Context) (*WineInfoResult, error) {
		prompt := buildWineInfoPrompt(key)
		resp, err := gp.model.GenerateContent(callCtx, genai.Text(prompt))
		if err != nil {
			return nil, fmt.Errorf("failed to generate content: %w", err)
		}

		raw, err := extractResponseText(resp)
		if err != nil {
			return nil, fmt.Errorf("failed to extract gemini response: %w", err)
		}

		return parseWineInfoResult(raw)
	})
}
