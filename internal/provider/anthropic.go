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

// Anthropic implements the Provider interface using the Anthropic Messages API.
type Anthropic struct {
	apiKey  string
	model   string
	baseURL string
}

func (a *Anthropic) Name() string { return "anthropic" }

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model string `json:"model"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

func (a *Anthropic) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	model := req.Model
	if model == "" {
		model = a.model
	}
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 8192
	}

	body := anthropicRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    req.SystemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: req.UserMessage},
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := strings.TrimRight(a.baseURL, "/") + "/v1/messages"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

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
		return nil, fmt.Errorf("anthropic API error (HTTP %d): %s", resp.StatusCode, string(respData))
	}

	var apiResp anthropicResponse
	if err := json.Unmarshal(respData, &apiResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("anthropic API error: %s: %s", apiResp.Error.Type, apiResp.Error.Message)
	}

	var content string
	for _, c := range apiResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &GenerateResponse{
		Content:   content,
		Model:     apiResp.Model,
		TokensIn:  apiResp.Usage.InputTokens,
		TokensOut: apiResp.Usage.OutputTokens,
	}, nil
}
