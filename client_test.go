package ternwire

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const testAPIKey = "em_test_origincheck"

func TestNewClientRejectsCleartextNonLoopbackOrigins(t *testing.T) {
	for _, baseURL := range []string{
		"http://api.example.com",
		"http://10.0.0.5:3001",
		"http://api:3001",
		"http://localhost.example.com",
		"http://127.0.0.1.nip.io",
		"http://[::2]",
	} {
		client, err := NewClient(Config{APIKey: testAPIKey, BaseURL: baseURL})
		if client != nil || err == nil || !strings.Contains(err.Error(), "must use https") {
			t.Errorf("%s: got client=%v err=%v, want an https error", baseURL, client, err)
		}
	}
}

func TestNewClientAcceptsHTTPSAndLoopbackOrigins(t *testing.T) {
	for _, baseURL := range []string{
		"https://api.example.com",
		"http://localhost:3001",
		"http://LOCALHOST:3001",
		"http://api.localhost:3001",
		"http://127.10.20.30",
		"http://[::1]:3001",
	} {
		if _, err := NewClient(Config{APIKey: testAPIKey, BaseURL: baseURL}); err != nil {
			t.Errorf("%s: %v", baseURL, err)
		}
	}
}

func TestLoopbackClientSendsTheKey(t *testing.T) {
	var authorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client, err := NewClient(Config{APIKey: testAPIKey, BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.DeleteEmailTemplate(context.Background(), "tpl_00000001"); err != nil {
		t.Fatal(err)
	}
	if authorization != "Bearer "+testAPIKey {
		t.Fatalf("Authorization = %q", authorization)
	}
}

func TestLoopbackClientIgnoresEnvironmentProxies(t *testing.T) {
	// net/http caches the proxy environment on first use, so run in a fresh process.
	if os.Getenv("TERNWIRE_PROXY_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestLoopbackClientIgnoresEnvironmentProxies$")
		cmd.Env = append(os.Environ(), "TERNWIRE_PROXY_CHILD=1")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v\n%s", err, output)
		}
		return
	}
	var apiAuthorizations, proxyAuthorizations []string
	page := []byte(`{"data":[],"page":{"nextCursor":null,"hasMore":false}}`)
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiAuthorizations = append(apiAuthorizations, r.Header.Get("Authorization"))
		_, _ = w.Write(page)
	}))
	defer api.Close()
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyAuthorizations = append(proxyAuthorizations, r.Header.Get("Authorization"))
		_, _ = w.Write(page)
	}))
	defer proxy.Close()
	for _, name := range []string{"http_proxy", "HTTP_PROXY", "all_proxy", "ALL_PROXY"} {
		t.Setenv(name, proxy.URL)
	}
	t.Setenv("no_proxy", "")
	t.Setenv("NO_PROXY", "")
	port := strings.TrimPrefix(api.URL, "http://127.0.0.1:")
	for _, host := range []string{"127.0.0.1", "localhost", "api.localhost"} {
		client, err := NewClient(Config{APIKey: testAPIKey, BaseURL: "http://" + host + ":" + port})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.ListEmailTemplates(context.Background(), ListEmailTemplatesQuery{}); err != nil {
			t.Fatalf("%s: %v", host, err)
		}
	}
	if len(proxyAuthorizations) != 0 {
		t.Fatalf("proxy received %q", proxyAuthorizations)
	}
	if len(apiAuthorizations) != 3 {
		t.Fatalf("API received %q", apiAuthorizations)
	}
}

func TestListEmailTemplatesSendsCursorAndLimit(t *testing.T) {
	var targets []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targets = append(targets, r.URL.RequestURI())
		_, _ = w.Write([]byte(`{"data":[],"page":{"nextCursor":null,"hasMore":false}}`))
	}))
	defer server.Close()
	client, err := NewClient(Config{APIKey: testAPIKey, BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.ListEmailTemplates(context.Background(), ListEmailTemplatesQuery{Cursor: "c+/= x", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListEmailTemplates(context.Background(), ListEmailTemplatesQuery{}); err != nil {
		t.Fatal(err)
	}
	if page, ok := result["page"].(Object); !ok || page["hasMore"] != false {
		t.Fatalf("result = %v", result)
	}
	want := []string{"/v1/templates?cursor=c%2B%2F%3D+x&limit=10", "/v1/templates"}
	if strings.Join(targets, " ") != strings.Join(want, " ") {
		t.Fatalf("targets = %q, want %q", targets, want)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestClientSendsRequestsToTheConfiguredOrigin(t *testing.T) {
	for _, tc := range []struct{ baseURL, want string }{
		{"", "https://ternwire.com/v1/templates"},
		{"https://staging.example.com/", "https://staging.example.com/v1/templates"},
	} {
		client, err := NewClient(Config{APIKey: testAPIKey, BaseURL: tc.baseURL})
		if err != nil {
			t.Fatalf("%q: %v", tc.baseURL, err)
		}
		var targets []string
		client.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
			targets = append(targets, r.URL.String())
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": {"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"data":[],"page":{"nextCursor":null,"hasMore":false}}`)),
				Request:    r,
			}, nil
		})
		if _, err := client.ListEmailTemplates(context.Background(), ListEmailTemplatesQuery{}); err != nil {
			t.Fatalf("%q: %v", tc.baseURL, err)
		}
		if len(targets) != 1 || targets[0] != tc.want {
			t.Errorf("%q: targets = %q, want [%q]", tc.baseURL, targets, tc.want)
		}
	}
}
