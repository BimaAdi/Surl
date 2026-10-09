package response

import (
	"testing"

	"github.com/BimaAdi/surl/core"
)

func TestRawResponseNonVerboseReturnsBodyOnly(t *testing.T) {
	resp := core.Response{
		Status: 200,
		Header: map[string]string{"Content-Type": "text/plain"},
		Body:   "hello",
	}
	got, err := RawResponse(resp, false)
	if err != nil {
		t.Fatalf("RawResponse() error = %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("RawResponse() = %q, want %q", got, "hello")
	}
}

func TestRawResponseVerboseIncludesSortedHeadersAndBody(t *testing.T) {
	resp := core.Response{
		Status: 200,
		Header: map[string]string{"X-B": "2", "Content-Type": "text/plain", "X-A": "1"},
		Body:   "hello",
	}
	got, err := RawResponse(resp, true)
	if err != nil {
		t.Fatalf("RawResponse() error = %v", err)
	}
	want := "status: 200\nContent-Type: text/plain\nX-A: 1\nX-B: 2\n\nhello"
	if string(got) != want {
		t.Fatalf("RawResponse() = %q, want %q", got, want)
	}
}

func TestRawResponseVerboseWithoutHeadersKeepsBlankLine(t *testing.T) {
	resp := core.Response{Body: "hello"}
	got, err := RawResponse(resp, true)
	if err != nil {
		t.Fatalf("RawResponse() error = %v", err)
	}
	if want := "status: 0\n\nhello"; string(got) != want {
		t.Fatalf("RawResponse() = %q, want %q", got, want)
	}
}

func TestRawResponseVerboseIncludesNonOKStatus(t *testing.T) {
	resp := core.Response{
		Status: 404,
		Body:   "not found",
	}
	got, err := RawResponse(resp, true)
	if err != nil {
		t.Fatalf("RawResponse() error = %v", err)
	}
	if want := "status: 404\n\nnot found"; string(got) != want {
		t.Fatalf("RawResponse() = %q, want %q", got, want)
	}
}

func TestRawResponseVerboseEmptyBody(t *testing.T) {
	resp := core.Response{
		Header: map[string]string{"X-A": "1"},
	}
	got, err := RawResponse(resp, true)
	if err != nil {
		t.Fatalf("RawResponse() error = %v", err)
	}
	if want := "status: 0\nX-A: 1\n\n"; string(got) != want {
		t.Fatalf("RawResponse() = %q, want %q", got, want)
	}
}

func TestRawResponseNonVerboseEmptyBody(t *testing.T) {
	got, err := RawResponse(core.Response{}, false)
	if err != nil {
		t.Fatalf("RawResponse() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("RawResponse() = %q, want empty", got)
	}
}
