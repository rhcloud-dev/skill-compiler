package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAI implements the Provider interface using the OpenAI Chat Completions API.
type OpenAI struct {
	apiKey  string
	model   string
	baseURL string
}

func (o *OpenAI) Name() string { return "openai" }

type openaiRequest struct {
	Model               string          `json:"model"`
	Messages            []openaiMessage `json:"messages"`
	MaxCompletionTokens int             `json:"max_completion_tokens,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func (o *OpenAI) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	model := req.Model
	if model == "" {
		model = o.model
	}

	messages := []openaiMessage{}
	if req.SystemPrompt != "" {
		messages = append(messages, openaiMessage{Role: "system", Content: req.SystemPrompt})
	}
	messages = append(messages, openaiMessage{Role: "user", Content: req.UserMessage})

	body := openaiRequest{
		Model:    model,
		Messages: messages,
	}
	if req.MaxTokens > 0 {
		body.MaxCompletionTokens = req.MaxTokens
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := strings.TrimRight(o.baseURL, "/") + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai API error (HTTP %d): %s", resp.StatusCode, string(respData))
	}

	var apiResp openaiResponse
	if err := json.Unmarshal(respData, &apiResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("openai API error: %s: %s", apiResp.Error.Type, apiResp.Error.Message)
	}

	content := ""
	if len(apiResp.Choices) > 0 {
		content = apiResp.Choices[0].Message.Content
	}

	return &GenerateResponse{
		Content:   content,
		Model:     apiResp.Model,
		TokensIn:  apiResp.Usage.PromptTokens,
		TokensOut: apiResp.Usage.CompletionTokens,
	}, nil
}
