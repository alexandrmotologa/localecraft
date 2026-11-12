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

// GeminiProvider handles communication with Google Gemini Generative Language API.
type GeminiProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewGeminiProvider constructs a GeminiProvider.
func NewGeminiProvider(apiKey string, model string) *GeminiProvider {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &GeminiProvider{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (g *GeminiProvider) Name() string {
	return "gemini"
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiRequest struct {
	Contents          []geminiContent `json:"contents"`
	SystemInstruction *geminiContent  `json:"system_instruction,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

func (g *GeminiProvider) Translate(ctx context.Context, req TranslateRequest) (TranslateResult, error) {
	if g.apiKey == "" {
		return TranslateResult{}, fmt.Errorf("gemini api key is required")
	}

	systemPrompt := `You are an expert software localization engineer and UI typography specialist.
Your task is to translate software user interface strings with extreme accuracy, preserving UI layout constraints and ICU MessageFormat syntax.

Rules:
1. Preserve all placeholders like {name}, {count}, {item} exactly without translating or reformatting them.
2. If the string contains ICU plural or select syntax, retain full syntax balance and valid CLDR plural keywords.
3. Use modern, natural native software terminology appropriate for high-growth tech products.
4. Output ONLY the translated string without quotes, conversational commentary, or markdown formatting.`

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
		if len(req.PreviousCandidates) > 0 {
			userPrompt.WriteString(fmt.Sprintf("Previous candidates that overflowed: %s\n", strings.Join(req.PreviousCandidates, ", ")))
		}
	}

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{
				Role:  "user",
				Parts: []geminiPart{{Text: userPrompt.String()}},
			},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return TranslateResult{}, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", g.model, g.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return TranslateResult{}, fmt.Errorf("create http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return TranslateResult{}, fmt.Errorf("call gemini api: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return TranslateResult{}, fmt.Errorf("read response body: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &geminiResp); err != nil {
		return TranslateResult{}, fmt.Errorf("decode gemini response: %w", err)
	}

	if geminiResp.Error != nil {
		return TranslateResult{}, fmt.Errorf("gemini api error (%d): %s", geminiResp.Error.Code, geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return TranslateResult{}, fmt.Errorf("no translation candidates returned from gemini")
	}

	translated := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	// Clean wrapping quotes if model included them
	translated = strings.TrimPrefix(translated, "\"")
	translated = strings.TrimSuffix(translated, "\"")

	return TranslateResult{
		Translation: translated,
		Explanation: fmt.Sprintf("Generated with Gemini %s", g.model),
		Provider:    g.Name(),
	}, nil
}
