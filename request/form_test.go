package request_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BimaAdi/surl/core"
	"github.com/BimaAdi/surl/request"
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

	_, err := core.ExecuteRequest(context.Background(), server.Client(), core.RequestConfig{
		URL:    server.URL,
		Method: http.MethodPost,
		Form: []request.FormValue{
			{Name: "title", Value: "some title"},
			{Name: "body", Value: "some body"},
		},
	}, core.Config{})
	if err != nil {
		t.Fatal(err)
	}
}
