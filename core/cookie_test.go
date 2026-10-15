package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecuteRequestAppliesGlobalCookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, err := r.Cookie("session_id")
		if err != nil || got.Value != "global-token" {
			t.Errorf("session_id cookie = %v, %v; want global-token", got, err)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	var cfg Config
	cfg.Global.Cookies = map[string]string{"session_id": "global-token"}
	result, err := ExecuteRequest(context.Background(), server.Client(), RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
	}, cfg)
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("ExecuteRequest() = %#v, %v", result, err)
	}
}

func TestExecuteRequestLocalCookiesOverrideGlobalCookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for name, want := range map[string]string{
			"session_id": "local-token",
			"theme":      "dark",
		} {
			cookie, err := r.Cookie(name)
			if err != nil || cookie.Value != want {
				t.Errorf("%s cookie = %v, %v; want %q", name, cookie, err, want)
			}
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	var cfg Config
	cfg.Global.Cookies = map[string]string{"session_id": "global-token", "theme": "dark"}
	_, err := ExecuteRequest(context.Background(), server.Client(), RequestConfig{
		URL:     server.URL,
		Method:  http.MethodGet,
		Cookies: map[string]string{"session_id": "local-token"},
	}, cfg)
	if err != nil {
		t.Fatal(err)
	}
}
