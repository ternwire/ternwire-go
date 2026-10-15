// Package ternwire provides a server-side client for the Ternwire public API.
package ternwire

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Object = map[string]any

type Config struct {
	APIKey  string
	BaseURL string
	Timeout time.Duration
}

type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

type APIError struct {
	Status  int
	RawBody string
	Body    any
	Headers http.Header
}

func (e *APIError) Error() string { return fmt.Sprintf("Ternwire API returned HTTP %d", e.Status) }

var apiKeyPattern = regexp.MustCompile(`^em_(live|test)_[A-Za-z0-9_-]+$`)
var resourceIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{7,127}$`)
var idempotencyPattern = regexp.MustCompile(`^[\x21-\x7e]{1,128}$`)

func NewClient(config Config) (*Client, error) {
	if !apiKeyPattern.MatchString(config.APIKey) {
		return nil, fmt.Errorf("use an em_live_ or em_test_ API key")
	}
	origin, err := url.Parse(config.BaseURL)
	if err != nil || origin == nil || (origin.Scheme != "https" && origin.Scheme != "http") ||
		origin.Hostname() == "" || origin.User != nil || (origin.Path != "" && origin.Path != "/") ||
		origin.RawQuery != "" || origin.ForceQuery || origin.Fragment != "" {
		return nil, fmt.Errorf("BaseURL must be an HTTP(S) origin without /v1")
	}
	loopback := isLoopbackHost(origin.Hostname())
	if origin.Scheme == "http" && !loopback {
		return nil, fmt.Errorf("BaseURL must use https unless it points at localhost or a loopback address")
	}
	if config.Timeout < 0 {
		return nil, fmt.Errorf("Timeout must be positive")
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	httpClient := &http.Client{
		Timeout:       config.Timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	if loopback {
		// A proxy cannot reach this machine's loopback, and would receive the API key.
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.Proxy = nil
		httpClient.Transport = transport
	}
	return &Client{apiKey: config.APIKey, baseURL: strings.TrimRight(config.BaseURL, "/"), http: httpClient}, nil
}

func isLoopbackHost(host string) bool {
	host = strings.ToLower(host)
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	addr, err := netip.ParseAddr(host)
	return err == nil && addr.IsLoopback()
}

// BEGIN GENERATED OPERATIONS
func (c *Client) SubmitEmail(ctx context.Context, input Object, idempotencyKey string) (Object, error) {
	if input == nil {
		return nil, fmt.Errorf("input must be a JSON object")
	}
	if !idempotencyPattern.MatchString(idempotencyKey) {
		return nil, fmt.Errorf("idempotencyKey must contain 1-128 visible ASCII characters")
	}
	return c.request(ctx, "POST", "/v1/emails", 202, input, idempotencyKey)
}

func (c *Client) GetMessage(ctx context.Context, messageId string) (Object, error) {
	if !resourceIDPattern.MatchString(messageId) {
		return nil, fmt.Errorf("invalid Ternwire resource ID")
	}
	return c.request(ctx, "GET", "/v1/emails/"+messageId, 200, nil, "")
}

func (c *Client) CreateEmailTemplate(ctx context.Context, input Object) (Object, error) {
	if input == nil {
		return nil, fmt.Errorf("input must be a JSON object")
	}
	return c.request(ctx, "POST", "/v1/templates", 201, input, "")
}

// ListEmailTemplatesQuery holds optional query parameters. Zero values are left out.
type ListEmailTemplatesQuery struct {
	Cursor string
	Limit  int
}

func (c *Client) ListEmailTemplates(ctx context.Context, query ListEmailTemplatesQuery) (Object, error) {
	values := url.Values{}
	if query.Cursor != "" {
		values.Set("cursor", query.Cursor)
	}
	if query.Limit != 0 {
		values.Set("limit", strconv.Itoa(query.Limit))
	}
	return c.request(ctx, "GET", "/v1/templates"+queryString(values), 200, nil, "")
}

func (c *Client) GetEmailTemplate(ctx context.Context, templateId string) (Object, error) {
	if !resourceIDPattern.MatchString(templateId) {
		return nil, fmt.Errorf("invalid Ternwire resource ID")
	}
	return c.request(ctx, "GET", "/v1/templates/"+templateId, 200, nil, "")
}

func (c *Client) ReplaceEmailTemplate(ctx context.Context, templateId string, input Object) (Object, error) {
	if !resourceIDPattern.MatchString(templateId) {
		return nil, fmt.Errorf("invalid Ternwire resource ID")
	}
	if input == nil {
		return nil, fmt.Errorf("input must be a JSON object")
	}
	return c.request(ctx, "PUT", "/v1/templates/"+templateId, 200, input, "")
}

func (c *Client) DeleteEmailTemplate(ctx context.Context, templateId string) (Object, error) {
	if !resourceIDPattern.MatchString(templateId) {
		return nil, fmt.Errorf("invalid Ternwire resource ID")
	}
	return c.request(ctx, "DELETE", "/v1/templates/"+templateId, 204, nil, "")
}

func (c *Client) CreateAttachmentUploadSlot(ctx context.Context, input Object) (Object, error) {
	if input == nil {
		return nil, fmt.Errorf("input must be a JSON object")
	}
	return c.request(ctx, "POST", "/v1/attachments", 201, input, "")
}

func (c *Client) GetAttachment(ctx context.Context, attachmentId string) (Object, error) {
	if !resourceIDPattern.MatchString(attachmentId) {
		return nil, fmt.Errorf("invalid Ternwire resource ID")
	}
	return c.request(ctx, "GET", "/v1/attachments/"+attachmentId, 200, nil, "")
}

func (c *Client) CompleteAttachmentUpload(ctx context.Context, attachmentId string) (Object, error) {
	if !resourceIDPattern.MatchString(attachmentId) {
		return nil, fmt.Errorf("invalid Ternwire resource ID")
	}
	return c.request(ctx, "POST", "/v1/attachments/"+attachmentId+"/complete", 202, nil, "")
}

func (c *Client) CreateSendingDomain(ctx context.Context, input Object) (Object, error) {
	if input == nil {
		return nil, fmt.Errorf("input must be a JSON object")
	}
	return c.request(ctx, "POST", "/v1/domains", 201, input, "")
}

func (c *Client) GetSendingDomain(ctx context.Context, domainId string) (Object, error) {
	if !resourceIDPattern.MatchString(domainId) {
		return nil, fmt.Errorf("invalid Ternwire resource ID")
	}
	return c.request(ctx, "GET", "/v1/domains/"+domainId, 200, nil, "")
}

func (c *Client) DeleteSendingDomain(ctx context.Context, domainId string) (Object, error) {
	if !resourceIDPattern.MatchString(domainId) {
		return nil, fmt.Errorf("invalid Ternwire resource ID")
	}
	return c.request(ctx, "DELETE", "/v1/domains/"+domainId, 202, nil, "")
}

func (c *Client) CreateWebhookEndpoint(ctx context.Context, input Object) (Object, error) {
	if input == nil {
		return nil, fmt.Errorf("input must be a JSON object")
	}
	return c.request(ctx, "POST", "/v1/webhook-endpoints", 201, input, "")
}

func (c *Client) RotateWebhookEndpointSigningSecret(ctx context.Context, endpointId string) (Object, error) {
	if !resourceIDPattern.MatchString(endpointId) {
		return nil, fmt.Errorf("invalid Ternwire resource ID")
	}
	return c.request(ctx, "POST", "/v1/webhook-endpoints/"+endpointId+"/rotate-signing-secret", 200, nil, "")
}

func (c *Client) GetWebhookAttempt(ctx context.Context, attemptId string) (Object, error) {
	if !resourceIDPattern.MatchString(attemptId) {
		return nil, fmt.Errorf("invalid Ternwire resource ID")
	}
	return c.request(ctx, "GET", "/v1/webhook-attempts/"+attemptId, 200, nil, "")
}

func (c *Client) ReplayWebhookAttempt(ctx context.Context, attemptId string) (Object, error) {
	if !resourceIDPattern.MatchString(attemptId) {
		return nil, fmt.Errorf("invalid Ternwire resource ID")
	}
	return c.request(ctx, "POST", "/v1/webhook-attempts/"+attemptId+"/replay", 202, nil, "")
}

// END GENERATED OPERATIONS

func queryString(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	return "?" + values.Encode()
}

func (c *Client) request(ctx context.Context, method, path string, expectedStatus int, body Object, idempotencyKey string) (Object, error) {
	var data io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		data = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	response, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != expectedStatus {
		apiError := &APIError{Status: response.StatusCode, RawBody: string(raw), Headers: response.Header.Clone()}
		_ = json.Unmarshal(raw, &apiError.Body)
		return nil, apiError
	}
	if expectedStatus == http.StatusNoContent {
		return nil, nil
	}
	var result Object
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("invalid Ternwire response: %w", err)
	}
	if result == nil {
		return nil, fmt.Errorf("Ternwire API returned a non-object JSON response")
	}
	return result, nil
}
