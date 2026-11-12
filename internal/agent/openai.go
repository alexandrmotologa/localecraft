package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIProvider handles communication with OpenAI or compatible local LLM endpoints (Ollama, vLLM).
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOpenAIProvider creates an OpenAIProvider.
func NewOpenAIProvider(apiKey string, baseURL string, model string) *OpenAIProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (o *OpenAIProvider) Name() string {
	return "openai"
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model       string              `json:"model"`
	Messages    []openAIChatMessage `json:"messages"`
	Temperature float64             `json:"temperature"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (o *OpenAIProvider) Translate(ctx context.Context, req TranslateRequest) (TranslateResult, error) {
	systemPrompt := `You are an expert software localization engineer and UI typography specialist.
Your task is to translate software user interface strings with extreme accuracy, preserving UI layout constraints and ICU MessageFormat syntax.

Rules:
1. Preserve all placeholders like {name}, {count}, {item} exactly without translating or altering them.
2. If the string contains ICU plural or select syntax, retain full syntax balance and valid CLDR plural keywords.
3. Output ONLY the translated string without quotes, conversational commentary, or markdown.`

	var userPrompt strings.Builder
	userPrompt.WriteString(fmt.Sprintf("Translate this UI string from '%s' to '%s'.\n", req.SourceLocale, req.TargetLocale))
	userPrompt.WriteString(fmt.Sprintf("Source Text: %s\n", req.SourceText))

	if req.Context != "" {
		userPrompt.WriteString(fmt.Sprintf("UI Container Context: %s\n", req.Context))
	}

	if len(req.Glossary) > 0 {
		userPrompt.WriteString("Mandatory Glossary:\n")
		for k, v := range req.Glossary {
			userPrompt.WriteString(fmt.Sprintf("- %s -> %s\n", k, v))
		}
	}

	if req.TargetCharLimit > 0 {
		userPrompt.WriteString(fmt.Sprintf("PHYSICAL CONSTRAINT: The target UI container will clip or overflow if text exceeds %d characters.\n", req.TargetCharLimit))
		userPrompt.WriteString(fmt.Sprintf("You MUST synthesize a natural, punchy synonym or concise phrasing that is %d characters or fewer.\n", req.TargetCharLimit))
	}

	reqPayload := openAIChatRequest{
		Model: o.model,
		Messages: []openAIChatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt.String()},
		},
		Temperature: 0.2,
	}

	payload, err := json.Marshal(reqPayload)
	if err != nil {
		return TranslateResult{}, fmt.Errorf("marshal request: %w", err)
	}

	endpoint := o.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return TranslateResult{}, fmt.Errorf("create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if o.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	resp, err := o.httpClient.Do(httpReq)
	if err != nil {
		return TranslateResult{}, fmt.Errorf("call openai api: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return TranslateResult{}, fmt.Errorf("read response body: %w", err)
	}

	var chatResp openAIChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return TranslateResult{}, fmt.Errorf("decode response: %w", err)
	}

	if chatResp.Error != nil {
		return TranslateResult{}, fmt.Errorf("api error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return TranslateResult{}, fmt.Errorf("no choices returned from api")
	}

	translated := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	translated = strings.TrimPrefix(translated, "\"")
	translated = strings.TrimSuffix(translated, "\"")

	return TranslateResult{
		Translation: translated,
		Explanation: fmt.Sprintf("Generated with %s (%s)", o.model, o.Name()),
		Provider:    o.Name(),
	}, nil
}
