package main

import (
	"context"
	"encoding/json"
	"github.com/BimaAdi/surl/core"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestExecuteRequestUsesMetadataAndVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/users" || r.URL.Query().Get("active") != "true" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		if got := r.Header.Get("X-Test"); got != "value" {
			t.Errorf("X-Test header = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"name":"value"}` {
			t.Errorf("body = %q", body)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	result, err := core.ExecuteRequest(context.Background(), server.Client(), requestConfig{
		URL:     "{base}/users",
		Method:  "post",
		Headers: map[string]string{"X-Test": "{value}"},
		Query:   map[string]string{"active": "true"},
		Body:    []byte(`{"name":"{value}"}`),
	}, map[string]string{"base": server.URL, "value": "value"})
	if err != nil || result.Status != http.StatusOK || result.Body != "ok" {
		t.Fatalf("executeRequest() = %#v, %v", result, err)
	}
}

func TestExecuteRequestDecodesStringBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"title":"some title"}` {
			t.Errorf("body = %q", body)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	_, err := core.ExecuteRequest(context.Background(), server.Client(), requestConfig{
		URL:    server.URL,
		Method: http.MethodPost,
		Body:   json.RawMessage(`"{\"title\":\"some title\"}"`),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExecuteRequestUsesJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"title":"some title","body":"some body"}` {
			t.Errorf("body = %q", body)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	_, err := core.ExecuteRequest(context.Background(), server.Client(), requestConfig{
		URL:    server.URL,
		Method: http.MethodPost,
		JSON:   json.RawMessage(`{"title":"some title","body":"some body"}`),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestExecuteRequestUsesFormBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type = %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `title=some+title&body=some+body` {
			t.Errorf("body = %q", body)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	_, err := core.ExecuteRequest(context.Background(), server.Client(), requestConfig{
		URL:    server.URL,
		Method: http.MethodPost,
		Form: []core.FormValue{
			{Name: "title", Value: "some title"},
			{Name: "body", Value: "some body"},
		},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

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
	if err := app.Run(context.Background(), []string{"surl", "--conf", configFile, "run", "get"}); err != nil {
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
