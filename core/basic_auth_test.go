package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecuteRequestAppliesGlobalBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "myusername" || password != "some-password" {
			t.Errorf("BasicAuth = %q, %q, %v", username, password, ok)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	var cfg Config
	cfg.Global.BasicAuth = &BasicAuthConfig{Username: "myusername", Password: "some-password"}
	result, err := ExecuteRequest(context.Background(), server.Client(), RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
	}, cfg)
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("ExecuteRequest() = %#v, %v", result, err)
	}
}

func TestExecuteRequestAPIBasicAuthOverridesGlobalBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "local-user" || password != "local-password" {
			t.Errorf("BasicAuth = %q, %q, %v", username, password, ok)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	var cfg Config
	cfg.Global.BasicAuth = &BasicAuthConfig{Username: "global-user", Password: "global-password"}
	reqConfig := RequestConfig{
		URL:       server.URL,
		Method:    http.MethodGet,
		BasicAuth: &BasicAuthConfig{Username: "local-user", Password: "local-password"},
	}
	result, err := ExecuteRequest(context.Background(), server.Client(), reqConfig, cfg)
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("ExecuteRequest() = %#v, %v", result, err)
	}
}
