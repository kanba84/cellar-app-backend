package llm

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

func extractResponseText(resp *genai.GenerateContentResponse) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("gemini response is nil")
	}
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in gemini response")
	}

	candidate := resp.Candidates[0]
	if candidate == nil {
		return "", fmt.Errorf("first gemini candidate is nil")
	}
	if candidate.Content == nil {
		return "", fmt.Errorf("empty content in gemini response")
	}
	if len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no parts in gemini response content")
	}

	var b strings.Builder
	for _, part := range candidate.Content.Parts {
		if part == nil {
			continue
		}
		if textPart, ok := part.(genai.Text); ok {
			b.WriteString(string(textPart))
		}
	}

	text := strings.TrimSpace(b.String())
	if text == "" {
		return "", fmt.Errorf("no text content in gemini response")
	}
	return text, nil
}

func cleanJSONResponse(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```JSON")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func parseWineInfoResult(raw string) (*WineInfoResult, error) {
	cleaned := cleanJSONResponse(raw)

	result := &WineInfoResult{}
	if err := json.Unmarshal([]byte(cleaned), result); err != nil {
		log.Printf("llm: failed to parse wine info JSON: %v; raw response: %q", err, raw)
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	ApplyGrapeNormalization(result.Grapes)
	return result, nil
}
