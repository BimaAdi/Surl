package request_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/BimaAdi/surl/core"
)

func TestExecuteRequestReadsBodyFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.txt")
	if err := os.WriteFile(path, []byte(`{"title":"some title"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"title":"some title"}` {
			t.Errorf("body = %q", body)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	var cfg core.Config
	result, err := core.ExecuteRequest(context.Background(), server.Client(), core.RequestConfig{
		URL:      server.URL,
		Method:   http.MethodPost,
		ReadFile: path,
	}, cfg)
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("executeRequest() = %#v, %v", result, err)
	}
}

func TestExecuteRequestReadFileSupportsVariables(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "body.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "hello" {
			t.Errorf("body = %q", body)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	var cfg core.Config
	cfg.Global.Variable = map[string]string{"dir": dir}
	spec, cfg, err := core.SubstituteRequest(core.RequestConfig{
		URL:      server.URL,
		Method:   http.MethodPost,
		ReadFile: "{dir}/body.txt",
	}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := core.ExecuteRequest(context.Background(), server.Client(), spec, cfg); err != nil {
		t.Fatal(err)
	}
}

func TestExecuteRequestReadFileMissingFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("unexpected request")
	}))
	defer server.Close()

	_, err := core.ExecuteRequest(context.Background(), server.Client(), core.RequestConfig{
		URL:      server.URL,
		Method:   http.MethodPost,
		ReadFile: filepath.Join(t.TempDir(), "missing.txt"),
	}, core.Config{})
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
