package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecuteRequestAppliesGlobalHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer sometoken" {
			t.Errorf("Authorization header = %q", got)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	var cfg Config
	cfg.Global.Variable = map[string]string{"token": "sometoken"}
	cfg.Global.Headers = map[string]string{"Authorization": "Bearer {token}"}
	result, err := ExecuteRequest(context.Background(), server.Client(), RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
	}, cfg)
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("ExecuteRequest() = %#v, %v", result, err)
	}
}

func TestExecuteRequestLocalHeaderOverridesGlobalHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "local-token" {
			t.Errorf("Authorization header = %q", got)
		}
		if got := r.Header.Get("X-Global"); got != "global" {
			t.Errorf("X-Global header = %q", got)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	var cfg Config
	cfg.Global.Headers = map[string]string{"Authorization": "global-token", "X-Global": "global"}
	_, err := ExecuteRequest(context.Background(), server.Client(), RequestConfig{
		URL:     server.URL,
		Method:  http.MethodGet,
		Headers: map[string]string{"Authorization": "local-token"},
	}, cfg)
	if err != nil {
		t.Fatal(err)
	}
}
