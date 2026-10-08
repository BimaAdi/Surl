package request_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BimaAdi/surl/core"
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

	var cfg core.Config
	result, err := core.ExecuteRequest(context.Background(), server.Client(), core.RequestConfig{
		URL:     server.URL + "/users",
		Method:  "post",
		Headers: map[string]string{"X-Test": "value"},
		Query:   map[string]string{"active": "true"},
		Body:    []byte(`{"name":"value"}`),
	}, cfg)
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

	_, err := core.ExecuteRequest(context.Background(), server.Client(), core.RequestConfig{
		URL:    server.URL,
		Method: http.MethodPost,
		Body:   json.RawMessage(`"{\"title\":\"some title\"}"`),
	}, core.Config{})
	if err != nil {
		t.Fatal(err)
	}
}
