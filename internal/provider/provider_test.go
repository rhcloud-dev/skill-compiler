package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/roberthamel/skill-compiler/internal/config"
)

func TestNew_Anthropic(t *testing.T) {
	p, err := New(&config.Resolved{
		Provider: "anthropic",
		APIKey:   "test-key",
		Model:    "claude-sonnet-4-6",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "anthropic" {
		t.Errorf("Name() = %q, want %q", p.Name(), "anthropic")
	}
}

func TestNew_OpenAI(t *testing.T) {
	p, err := New(&config.Resolved{
		Provider: "openai",
		APIKey:   "test-key",
		Model:    "gpt-4o",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "openai" {
		t.Errorf("Name() = %q, want %q", p.Name(), "openai")
	}
}

func TestNew_MissingKey(t *testing.T) {
	_, err := New(&config.Resolved{Provider: "anthropic"})
	if err == nil {
		t.Fatal("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "API key required") {
		t.Errorf("error = %q, want to contain %q", err.Error(), "API key required")
	}
}

func TestNew_DefaultAnthropic(t *testing.T) {
	// No provider specified, no base URL → should default to anthropic
	p, err := New(&config.Resolved{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "anthropic" {
		t.Errorf("default provider should be anthropic, got %q", p.Name())
	}
}

func TestAnthropic_Generate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request format
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/v1/messages") {
			t.Errorf("path = %q, want /v1/messages", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("x-api-key = %q, want %q", r.Header.Get("x-api-key"), "test-key")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("anthropic-version = %q, want %q", r.Header.Get("anthropic-version"), "2023-06-01")
		}

		// Verify request body
		var req anthropicRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request: %v", err)
		}
		if req.Model != "test-model" {
			t.Errorf("model = %q, want %q", req.Model, "test-model")
		}
		if req.System != "system prompt" {
			t.Errorf("system = %q, want %q", req.System, "system prompt")
		}
		if len(req.Messages) != 1 || req.Messages[0].Content != "user message" {
			t.Errorf("messages = %+v, want single user message", req.Messages)
		}

		// Respond
		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: "response content"}},
			Model: "test-model",
		}
		resp.Usage.InputTokens = 10
		resp.Usage.OutputTokens = 20
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	prov := &Anthropic{apiKey: "test-key", model: "test-model", baseURL: server.URL}
	resp, err := prov.Generate(context.Background(), GenerateRequest{
		SystemPrompt: "system prompt",
		UserMessage:  "user message",
		MaxTokens:    1000,
	})
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}
	if resp.Content != "response content" {
		t.Errorf("content = %q, want %q", resp.Content, "response content")
	}
	if resp.TokensIn != 10 || resp.TokensOut != 20 {
		t.Errorf("tokens = %d/%d, want 10/20", resp.TokensIn, resp.TokensOut)
	}
}

func TestOpenAI_Generate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/v1/chat/completions") {
			t.Errorf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("auth = %q, want %q", r.Header.Get("Authorization"), "Bearer test-key")
		}

		var req openaiRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decoding request: %v", err)
		}
		if req.Model != "test-model" {
			t.Errorf("model = %q, want %q", req.Model, "test-model")
		}
		// Should have system + user messages
		if len(req.Messages) != 2 {
			t.Errorf("got %d messages, want 2", len(req.Messages))
		}

		resp := openaiResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{{Message: struct {
				Content string `json:"content"`
			}{Content: "openai response"}}},
			Model: "test-model",
		}
		resp.Usage.PromptTokens = 15
		resp.Usage.CompletionTokens = 25
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	prov := &OpenAI{apiKey: "test-key", model: "test-model", baseURL: server.URL}
	resp, err := prov.Generate(context.Background(), GenerateRequest{
		SystemPrompt: "system",
		UserMessage:  "user",
		MaxTokens:    1000,
	})
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}
	if resp.Content != "openai response" {
		t.Errorf("content = %q, want %q", resp.Content, "openai response")
	}
	if resp.TokensIn != 15 || resp.TokensOut != 25 {
		t.Errorf("tokens = %d/%d, want 15/25", resp.TokensIn, resp.TokensOut)
	}
}
