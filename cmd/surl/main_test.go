package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BimaAdi/surl/core"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSubstituteLeavesUnknownVariables(t *testing.T) {
	if got := core.Substitute("{known}/{unknown}", map[string]string{"known": "ok"}); got != "ok/{unknown}" {
		t.Errorf("substitute() = %q", got)
	}
}

func TestNewAppRunsConfiguredRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("response"))
	}))
	defer server.Close()
	configJSON := `{"global":{"variable":{"url":"` + server.URL + `"}},"api":{"get":{"url":"{url}","method":"GET"}}}`
	configFile := t.TempDir() + "/surl.json"
	if err := os.WriteFile(configFile, []byte(configJSON), 0600); err != nil {
		t.Fatal(err)
	}

	var output strings.Builder
	app := newApp()
	app.Writer = &output
	if err := app.Run(context.Background(), []string{"surl", "--conf", configFile, "run", "--output", "json", "--verbose", "get"}); err != nil {
		t.Fatal(err)
	}
	var result response
	if err := json.Unmarshal([]byte(output.String()), &result); err != nil {
		t.Fatalf("invalid response JSON: %v", err)
	}
	if result.Status != http.StatusOK || result.Body != "response" || result.Header["Content-Type"] != "text/plain; charset=utf-8" {
		t.Errorf("response = %#v", result)
	}
}

func TestNewAppWritesPrettifiedJSONToFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"title":"hello","body":"world"}`))
	}))
	defer server.Close()

	dir := t.TempDir()
	outputFile := filepath.Join(dir, "response.json")
	configJSON := fmt.Sprintf(`{"global":{"variable":{"url":%q,"outdir":%q}},"api":{"write":{"url":"{url}","method":"GET","writefile":"{outdir}/response.json","prettify-json":true}}}`, server.URL, dir)
	configFile := filepath.Join(dir, "surl.json")
	if err := os.WriteFile(configFile, []byte(configJSON), 0600); err != nil {
		t.Fatal(err)
	}

	var output strings.Builder
	app := newApp()
	app.Writer = &output
	if err := app.Run(context.Background(), []string{"surl", "--conf", configFile, "run", "write"}); err != nil {
		t.Fatal(err)
	}
	if output.String() != "" {
		t.Errorf("terminal output = %q, want empty", output.String())
	}
	got, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"title\": \"hello\",\n  \"body\": \"world\"\n}"
	if string(got) != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}

func TestNewAppListsConfiguredAPIs(t *testing.T) {
	configJSON := `{"api":{"zulu":{"url":"https://example.com"},"alpha":{"url":"https://example.com"}}}`
	configFile := t.TempDir() + "/surl.json"
	if err := os.WriteFile(configFile, []byte(configJSON), 0600); err != nil {
		t.Fatal(err)
	}

	var output strings.Builder
	app := newApp()
	app.Writer = &output
	if err := app.Run(context.Background(), []string{"surl", "list", "--conf", configFile}); err != nil {
		t.Fatal(err)
	}

	if got, want := output.String(), "alpha\nzulu\n"; got != want {
		t.Errorf("list output = %q, want %q", got, want)
	}
}
