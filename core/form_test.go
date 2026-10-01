package core

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

	_, err := ExecuteRequest(context.Background(), server.Client(), RequestConfig{
		URL:    server.URL,
		Method: http.MethodPost,
		Form: []FormValue{
			{Name: "title", Value: "some title"},
			{Name: "body", Value: "some body"},
		},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}
