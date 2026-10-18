package core

import (
	"context"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

// writeServerCertPEM writes the server certificate as PEM into a temporary
// file and returns its path.
func writeServerCertPEM(t *testing.T, server *httptest.Server) string {
	t.Helper()
	cert := server.Certificate()
	if cert == nil {
		t.Fatal("test server has no certificate")
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	path := filepath.Join(t.TempDir(), "ca.crt")
	if err := os.WriteFile(path, certPEM, 0o600); err != nil {
		t.Fatalf("write certificate file: %v", err)
	}
	return path
}

func TestExecuteRequestCertFilepathTrustsServer(t *testing.T) {
	server, client := requestTLSServer(t)
	certFilepath := writeServerCertPEM(t, server)

	reqConfig := RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
		TLSConfig: &TLSConfig{
			InsecureSkipVerify: false,
			CertFilepath:       &certFilepath,
		},
	}
	result, err := ExecuteRequest(context.Background(), client, reqConfig, Config{})
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("ExecuteRequest() = %#v, %v", result, err)
	}
}

func TestExecuteRequestCertFilepathInvalidPath(t *testing.T) {
	server, client := requestTLSServer(t)
	certFilepath := filepath.Join(t.TempDir(), "missing.crt")

	reqConfig := RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
		TLSConfig: &TLSConfig{
			CertFilepath: &certFilepath,
		},
	}
	if _, err := ExecuteRequest(context.Background(), client, reqConfig, Config{}); err == nil {
		t.Fatal("expected error for missing certificate file")
	}
}

func TestExecuteRequestCertFilepathInvalidContent(t *testing.T) {
	server, client := requestTLSServer(t)
	certFilepath := filepath.Join(t.TempDir(), "invalid.crt")
	if err := os.WriteFile(certFilepath, []byte("this is not a certificate"), 0o600); err != nil {
		t.Fatalf("write certificate file: %v", err)
	}

	reqConfig := RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
		TLSConfig: &TLSConfig{
			CertFilepath: &certFilepath,
		},
	}
	if _, err := ExecuteRequest(context.Background(), client, reqConfig, Config{}); err == nil {
		t.Fatal("expected error for invalid certificate file content")
	}
}

func TestExecuteRequestTLSVersionTLS12(t *testing.T) {
	server, client := requestTLSServer(t)
	tlsVersion := "tls12"

	reqConfig := RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
		TLSConfig: &TLSConfig{
			InsecureSkipVerify: true,
			TLSVersion:         &tlsVersion,
		},
	}
	result, err := ExecuteRequest(context.Background(), client, reqConfig, Config{})
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("ExecuteRequest() = %#v, %v", result, err)
	}
}

func TestExecuteRequestTLSVersionTLS13(t *testing.T) {
	server, client := requestTLSServer(t)
	tlsVersion := "tls13"

	reqConfig := RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
		TLSConfig: &TLSConfig{
			InsecureSkipVerify: true,
			TLSVersion:         &tlsVersion,
		},
	}
	result, err := ExecuteRequest(context.Background(), client, reqConfig, Config{})
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("ExecuteRequest() = %#v, %v", result, err)
	}
}

func TestExecuteRequestTLSVersionDefaultsToTLS13(t *testing.T) {
	server, client := requestTLSServer(t)

	// The httptest server speaks TLS 1.3, so an omitted tls-version (which
	// defaults to TLS 1.3) must succeed.
	reqConfig := RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
		TLSConfig: &TLSConfig{
			InsecureSkipVerify: true,
		},
	}
	result, err := ExecuteRequest(context.Background(), client, reqConfig, Config{})
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("ExecuteRequest() = %#v, %v", result, err)
	}
}

func TestExecuteRequestInvalidTLSVersion(t *testing.T) {
	server, client := requestTLSServer(t)
	tlsVersion := "ssl3"

	reqConfig := RequestConfig{
		URL:    server.URL,
		Method: http.MethodGet,
		TLSConfig: &TLSConfig{
			InsecureSkipVerify: true,
			TLSVersion:         &tlsVersion,
		},
	}
	_, err := ExecuteRequest(context.Background(), client, reqConfig, Config{})
	if err == nil {
		t.Fatal("expected error for invalid tls-version")
	}
}
