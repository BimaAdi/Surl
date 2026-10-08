package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// requestTLSServer returns a TLS server with a self-signed certificate and a
// plain HTTP client that does not trust it.
func requestTLSServer(t *testing.T) (*httptest.Server, *http.Client) {
	t.Helper()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(server.Close)
	return server, &http.Client{}
}

func TestExecuteRequestGlobalInsecureSkipVerify(t *testing.T) {
	server, client := requestTLSServer(t)

	var cfg Config
	cfg.Global.TLSConfig = &TLSConfig{InsecureSkipVerify: true}
	result, err := ExecuteRequest(context.Background(), client, RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
	}, cfg)
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("ExecuteRequest() = %#v, %v", result, err)
	}
}

func TestExecuteRequestLocalTLSConfigOverridesGlobal(t *testing.T) {
	server, client := requestTLSServer(t)

	var cfg Config
	cfg.Global.TLSConfig = &TLSConfig{InsecureSkipVerify: false}
	reqConfig := RequestConfig{
		URL:       server.URL,
		Method:    http.MethodGet,
		TLSConfig: &TLSConfig{InsecureSkipVerify: true},
	}
	result, err := ExecuteRequest(context.Background(), client, reqConfig, cfg)
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("ExecuteRequest() = %#v, %v", result, err)
	}
}

func TestExecuteRequestWithoutTLSConfigFailsOnUntrustedCert(t *testing.T) {
	server, client := requestTLSServer(t)

	_, err := ExecuteRequest(context.Background(), client, RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
	}, Config{})
	if err == nil {
		t.Fatal("expected request against untrusted certificate to fail")
	}
}
