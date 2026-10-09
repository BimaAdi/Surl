package response

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/BimaAdi/surl/core"
)

func compactJSON(t *testing.T, data []byte) string {
	t.Helper()
	var buf bytes.Buffer
	if err := json.Compact(&buf, data); err != nil {
		t.Fatalf("invalid JSON %q: %v", data, err)
	}
	return buf.String()
}

func TestEncodeJSONBodyOnlyEmbedsJSONBody(t *testing.T) {
	resp := core.Response{
		Status: 200,
		Header: map[string]string{"Content-Type": "application/json; charset=utf-8"},
		Body:   `{"name":"surl"}`,
	}
	got, err := ResponseJSON(resp, false)
	if err != nil {
		t.Fatalf("EncodeJSON() error = %v", err)
	}
	if want := `{"name":"surl"}`; compactJSON(t, got) != want {
		t.Fatalf("EncodeJSON() = %s, want %s", got, want)
	}
}

func TestEncodeJSONBodyOnlyHeaderNameIsCaseInsensitive(t *testing.T) {
	resp := core.Response{
		Header: map[string]string{"content-type": "Application/JSON"},
		Body:   `[1,2,3]`,
	}
	got, err := ResponseJSON(resp, false)
	if err != nil {
		t.Fatalf("EncodeJSON() error = %v", err)
	}
	if want := `[1,2,3]`; compactJSON(t, got) != want {
		t.Fatalf("EncodeJSON() = %s, want %s", got, want)
	}
}

func TestEncodeJSONBodyOnlyInvalidJSONBodyEncodesAsString(t *testing.T) {
	resp := core.Response{
		Header: map[string]string{"Content-Type": "application/json"},
		Body:   `{not json`,
	}
	got, err := ResponseJSON(resp, false)
	if err != nil {
		t.Fatalf("EncodeJSON() error = %v", err)
	}
	if want := `"{not json"`; compactJSON(t, got) != want {
		t.Fatalf("EncodeJSON() = %s, want %s", got, want)
	}
}

func TestEncodeJSONBodyOnlyMissingContentTypeEncodesBodyAsString(t *testing.T) {
	resp := core.Response{Body: `{"a":1}`}
	got, err := ResponseJSON(resp, false)
	if err != nil {
		t.Fatalf("EncodeJSON() error = %v", err)
	}
	if want := `"{\"a\":1}"`; compactJSON(t, got) != want {
		t.Fatalf("EncodeJSON() = %s, want %s", got, want)
	}
}

func TestEncodeJSONVerboseIncludesStatusHeaderAndBody(t *testing.T) {
	resp := core.Response{
		Status: 201,
		Header: map[string]string{"Content-Type": "application/json", "X-Id": "42"},
		Body:   `{"ok":true}`,
	}
	got, err := ResponseJSON(resp, true)
	if err != nil {
		t.Fatalf("EncodeJSON() error = %v", err)
	}
	want := `{"status":201,"header":{"Content-Type":"application/json","X-Id":"42"},"body":{"ok":true}}`
	if compactJSON(t, got) != want {
		t.Fatalf("EncodeJSON() = %s, want %s", got, want)
	}
}

func TestEncodeJSONVerboseNonJSONBodyIsString(t *testing.T) {
	resp := core.Response{
		Status: 500,
		Header: map[string]string{"Content-Type": "text/plain"},
		Body:   "boom",
	}
	got, err := ResponseJSON(resp, true)
	if err != nil {
		t.Fatalf("EncodeJSON() error = %v", err)
	}
	want := `{"status":500,"header":{"Content-Type":"text/plain"},"body":"boom"}`
	if compactJSON(t, got) != want {
		t.Fatalf("EncodeJSON() = %s, want %s", got, want)
	}
}

func TestEncodeJSONOutputIsIndented(t *testing.T) {
	resp := core.Response{
		Status: 200,
		Header: map[string]string{"Content-Type": "text/plain"},
		Body:   "hi",
	}
	got, err := ResponseJSON(resp, true)
	if err != nil {
		t.Fatalf("EncodeJSON() error = %v", err)
	}
	if want := "{\n  \"status\": 200,\n  \"header\": {\n    \"Content-Type\": \"text/plain\"\n  },\n  \"body\": \"hi\"\n}"; string(got) != want {
		t.Fatalf("EncodeJSON() = %q, want %q", got, want)
	}
}
