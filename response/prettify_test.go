package response

import "testing"

func TestPrettifyJSONIndentsValidJSON(t *testing.T) {
	got := PrettifyJSON([]byte(`{"title":"hello","tags":["a","b"]}`))
	want := "{\n  \"title\": \"hello\",\n  \"tags\": [\n    \"a\",\n    \"b\"\n  ]\n}"
	if string(got) != want {
		t.Fatalf("PrettifyJSON() = %q, want %q", got, want)
	}
}

func TestPrettifyJSONKeepsTrailingNewline(t *testing.T) {
	got := PrettifyJSON([]byte("{\"a\":1}\n"))
	want := "{\n  \"a\": 1\n}\n"
	if string(got) != want {
		t.Fatalf("PrettifyJSON() = %q, want %q", got, want)
	}
}

func TestPrettifyJSONLeavesNonJSONUnchanged(t *testing.T) {
	in := []byte("not json")
	if got := PrettifyJSON(in); string(got) != string(in) {
		t.Fatalf("PrettifyJSON() = %q, want %q", got, in)
	}
}

func TestPrettifyJSONLeavesPlainTextUnchanged(t *testing.T) {
	in := []byte("status: 200\n\nOK")
	if got := PrettifyJSON(in); string(got) != string(in) {
		t.Fatalf("PrettifyJSON() = %q, want %q", got, in)
	}
}
