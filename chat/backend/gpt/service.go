package gpt

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

func CallOpenAIStructOutPut(
	ctx context.Context,
	apiKey, model string,
	messages []Message,
	responseFormat ResponseFormat,
) (*GPTResp, error) {
	url := "https://api.openai.com/v1/chat/completions"
	store := false

	payload := OpenAIRequest{
		Model:          model,
		Messages:       messages,
		ResponseFormat: &responseFormat,
		Store:          &store,
	}

	return callOpenAI(ctx, url, apiKey, payload)
}

func CallOpenAIChat(
	ctx context.Context,
	apiKey, model string,
	messages []Message,
) (*GPTResp, error) {
	url := "https://api.openai.com/v1/chat/completions"
	store := false

	payload := OpenAIRequest{
		Model:    model,
		Messages: messages,
		Store:    &store,
	}

	return callOpenAI(ctx, url, apiKey, payload)
}

func callOpenAI(ctx context.Context, url, apiKey string, payload OpenAIRequest) (*GPTResp, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao codificar payload OpenAI: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("erro ao criar requisição OpenAI: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 45 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar OpenAI: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta OpenAI: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("OpenAI retornou status %d: %s", resp.StatusCode, truncateBody(body, 500))
	}

	var apiResp GPTResp
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("erro ao interpretar JSON da OpenAI: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI não retornou respostas")
	}

	return &apiResp, nil
}

func truncateBody(body []byte, limit int) string {
	text := strings.TrimSpace(string(body))
	if len(text) <= limit {
		return text
	}
	return text[:limit] + "..."
}
