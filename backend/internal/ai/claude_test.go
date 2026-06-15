package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

const testModel = "claude-opus-4-8"

// decodeRequest reads and unmarshals the JSON request body the provider sent.
func decodeRequest(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	return payload
}

func TestClaudeProvider_Summarize_SendsExpectedRequestAndReturnsText(t *testing.T) {
	const userText = "We surveyed 200 customers and 80% asked for dark mode."

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/messages" {
			t.Errorf("path = %s, want /v1/messages", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "secret-key" {
			t.Errorf("x-api-key = %q, want %q", got, "secret-key")
		}
		if got := r.Header.Get("anthropic-version"); got != "2023-06-01" {
			t.Errorf("anthropic-version = %q, want 2023-06-01", got)
		}
		if got := r.Header.Get("content-type"); got != "application/json" {
			t.Errorf("content-type = %q, want application/json", got)
		}

		payload := decodeRequest(t, r)
		if payload["model"] != testModel {
			t.Errorf("model = %v, want %v", payload["model"], testModel)
		}
		messages, ok := payload["messages"].([]any)
		if !ok || len(messages) != 1 {
			t.Fatalf("messages = %v, want one message", payload["messages"])
		}
		first, _ := messages[0].(map[string]any)
		if first["role"] != "user" {
			t.Errorf("message role = %v, want user", first["role"])
		}
		if first["content"] != userText {
			t.Errorf("message content = %v, want %q", first["content"], userText)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"Most customers want dark mode."}]}`)
	}))
	defer srv.Close()

	p := NewClaudeProvider("secret-key", testModel, srv.URL, srv.Client())

	got, err := p.Summarize(context.Background(), userText)
	if err != nil {
		t.Fatalf("Summarize: unexpected error: %v", err)
	}
	if got != "Most customers want dark mode." {
		t.Errorf("Summarize = %q, want %q", got, "Most customers want dark mode.")
	}
}

func TestClaudeProvider_DetectBias_ParsesSummaryAndBiases(t *testing.T) {
	const canned = "Two biases detected.\n" +
		"- Anchoring: relied on first estimate.\n" +
		"- Confirmation bias: ignored disconfirming evidence."

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		body, _ := json.Marshal(map[string]any{
			"content": []map[string]any{{"type": "text", "text": canned}},
		})
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	p := NewClaudeProvider("secret-key", testModel, srv.URL, srv.Client())

	report, err := p.DetectBias(context.Background(), "some reasoning")
	if err != nil {
		t.Fatalf("DetectBias: unexpected error: %v", err)
	}
	if report.Summary != "Two biases detected." {
		t.Errorf("Summary = %q, want %q", report.Summary, "Two biases detected.")
	}
	if len(report.Biases) != 2 {
		t.Fatalf("len(Biases) = %d, want 2", len(report.Biases))
	}
	if report.Biases[0] != (DetectedBias{Name: "Anchoring", Explanation: "relied on first estimate."}) {
		t.Errorf("Biases[0] = %+v", report.Biases[0])
	}
	if report.Biases[1] != (DetectedBias{Name: "Confirmation bias", Explanation: "ignored disconfirming evidence."}) {
		t.Errorf("Biases[1] = %+v", report.Biases[1])
	}
}

func TestClaudeProvider_ConcatenatesTextBlocks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"content":[{"type":"text","text":"part one. "},{"type":"text","text":"part two."}]}`)
	}))
	defer srv.Close()

	p := NewClaudeProvider("secret-key", testModel, srv.URL, srv.Client())

	got, err := p.Critique(context.Background(), "an assumption")
	if err != nil {
		t.Fatalf("Critique: unexpected error: %v", err)
	}
	if got != "part one. part two." {
		t.Errorf("Critique = %q, want %q", got, "part one. part two.")
	}
}

func TestClaudeProvider_Non2xx_ReturnsInternalErrorWithoutAPIKey(t *testing.T) {
	const apiKey = "super-secret-key-value"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"error":{"message":"boom"}}`)
	}))
	defer srv.Close()

	p := NewClaudeProvider(apiKey, testModel, srv.URL, srv.Client())

	_, err := p.Summarize(context.Background(), "text")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if got := errors.KindOf(err); got != errors.KindInternal {
		t.Errorf("kind = %v, want KindInternal", got)
	}
	if got := errors.CodeOf(err); got != "AI_PROVIDER_ERROR" {
		t.Errorf("code = %q, want AI_PROVIDER_ERROR", got)
	}
	if msg := err.Error(); !strings.Contains(msg, "500") {
		t.Errorf("error message %q should include the status code 500", msg)
	}
	if msg := err.Error(); strings.Contains(msg, apiKey) {
		t.Errorf("error message %q must not contain the api key", msg)
	}
}
