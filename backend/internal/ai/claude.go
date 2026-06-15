package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/michaelgavalas/decision-intelligence-os/backend/pkg/errors"
)

// Anthropic Messages API integration constants. These mirror the vendor's
// published request contract for the /v1/messages endpoint.
const (
	defaultBaseURL    = "https://api.anthropic.com"
	anthropicVersion  = "2023-06-01"
	maxResponseTokens = 1024
	requestTimeout    = 30 * time.Second
)

// Task-specific system prompts. Each one instructs the model on how to behave
// for a single AI-assistance feature; the user's text is passed separately as
// the message content.
const (
	summarizeSystemPrompt = "You assist with decision-making. Write a concise, neutral " +
		"summary of the provided evidence. Do not add opinions, recommendations, or " +
		"information that is not present in the evidence."

	critiqueSystemPrompt = "You assist with decision-making. Critically evaluate the " +
		"provided assumption. Identify its strengths and weaknesses, and describe what " +
		"evidence would meaningfully increase or decrease confidence in it."

	detectBiasSystemPrompt = "You assist with decision-making by detecting cognitive " +
		"biases. On the first line, write a one-sentence summary. Then, on each " +
		"subsequent line, list one detected bias formatted exactly as " +
		"\"- <BiasName>: <explanation>\". If you detect no biases, write only the " +
		"summary line."
)

// ClaudeProvider implements LLMProvider by calling Anthropic's Messages API. It
// integrates the vendor as a product dependency, the same way an application
// integrates a payment or email provider.
type ClaudeProvider struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// NewClaudeProvider builds a provider. baseURL defaults to the public Anthropic
// endpoint when empty; httpClient defaults to a client with a sensible timeout.
func NewClaudeProvider(apiKey, model, baseURL string, httpClient *http.Client) *ClaudeProvider {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: requestTimeout}
	}
	return &ClaudeProvider{
		apiKey:     apiKey,
		model:      model,
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// messageRequest is the JSON body sent to the Messages API.
type messageRequest struct {
	Model     string           `json:"model"`
	MaxTokens int              `json:"max_tokens"`
	System    string           `json:"system"`
	Messages  []messageContent `json:"messages"`
}

// messageContent is a single conversational turn in the request.
type messageContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// messageResponse is the subset of the Messages API response this integration
// reads: a list of content blocks, of which the text blocks carry the answer.
type messageResponse struct {
	Content []contentBlock `json:"content"`
}

// contentBlock is one block of the model's response.
type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Summarize asks the model for a concise, neutral summary of the evidence.
func (p *ClaudeProvider) Summarize(ctx context.Context, text string) (string, error) {
	return p.complete(ctx, summarizeSystemPrompt, text)
}

// Critique asks the model for a critical evaluation of the assumption.
func (p *ClaudeProvider) Critique(ctx context.Context, statement string) (string, error) {
	return p.complete(ctx, critiqueSystemPrompt, statement)
}

// DetectBias asks the model to identify cognitive biases in the text and parses
// the structured response into a BiasReport.
func (p *ClaudeProvider) DetectBias(ctx context.Context, text string) (BiasReport, error) {
	out, err := p.complete(ctx, detectBiasSystemPrompt, text)
	if err != nil {
		return BiasReport{}, err
	}
	return parseBiasReport(out), nil
}

// complete performs one Messages API request with the given system prompt and
// user text, returning the concatenated text of the response.
func (p *ClaudeProvider) complete(ctx context.Context, systemPrompt, userText string) (string, error) {
	reqBody, err := json.Marshal(messageRequest{
		Model:     p.model,
		MaxTokens: maxResponseTokens,
		System:    systemPrompt,
		Messages:  []messageContent{{Role: "user", Content: userText}},
	})
	if err != nil {
		return "", errors.Wrap(err, errors.KindInternal, "AI_PROVIDER_ERROR", "failed to encode request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/messages", bytes.NewReader(reqBody))
	if err != nil {
		return "", errors.Wrap(err, errors.KindInternal, "AI_PROVIDER_ERROR", "failed to build request")
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", anthropicVersion)
	req.Header.Set("content-type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		// The underlying error may include the request URL but never the api key,
		// which is carried only in a header.
		return "", errors.Wrap(err, errors.KindInternal, "AI_PROVIDER_ERROR", "ai provider request failed")
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, errors.KindInternal, "AI_PROVIDER_ERROR", "failed to read ai provider response")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Include the status code for diagnostics; never include the api key.
		return "", errors.Internal(
			"AI_PROVIDER_ERROR",
			fmt.Sprintf("ai provider returned status %d", resp.StatusCode),
		)
	}

	var parsed messageResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", errors.Wrap(err, errors.KindInternal, "AI_PROVIDER_ERROR", "failed to decode ai provider response")
	}

	var sb strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return sb.String(), nil
}

// parseBiasReport interprets the model's structured bias output: the first
// non-empty line is the summary, and each subsequent line of the form
// "- Name: explanation" becomes a DetectedBias.
func parseBiasReport(out string) BiasReport {
	var report BiasReport
	summarySet := false

	for _, raw := range strings.Split(out, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		if !summarySet {
			report.Summary = line
			summarySet = true
			continue
		}
		if bias, ok := parseBiasLine(line); ok {
			report.Biases = append(report.Biases, bias)
		}
	}
	return report
}

// parseBiasLine parses a single "- Name: explanation" line into a DetectedBias.
// It reports false for lines that do not match the expected shape.
func parseBiasLine(line string) (DetectedBias, bool) {
	rest, ok := strings.CutPrefix(line, "-")
	if !ok {
		return DetectedBias{}, false
	}
	name, explanation, ok := strings.Cut(rest, ":")
	if !ok {
		return DetectedBias{}, false
	}
	name = strings.TrimSpace(name)
	explanation = strings.TrimSpace(explanation)
	if name == "" || explanation == "" {
		return DetectedBias{}, false
	}
	return DetectedBias{Name: name, Explanation: explanation}, true
}
