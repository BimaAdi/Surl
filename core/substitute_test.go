package core

import (
	"encoding/json"
	"testing"
)

func TestSubstitute(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		variables map[string]string
		want      string
	}{
		{
			name:      "no placeholders",
			value:     "plain text",
			variables: map[string]string{"name": "world"},
			want:      "plain text",
		},
		{
			name:      "single placeholder",
			value:     "hello {name}",
			variables: map[string]string{"name": "world"},
			want:      "hello world",
		},
		{
			name:      "repeated and multiple placeholders",
			value:     "{a}/{a}/{b}",
			variables: map[string]string{"a": "1", "b": "2"},
			want:      "1/1/2",
		},
		{
			name:      "unknown variable is kept",
			value:     "{missing}",
			variables: map[string]string{"name": "world"},
			want:      "{missing}",
		},
		{
			name:      "nil variables keeps placeholders",
			value:     "{a}",
			variables: nil,
			want:      "{a}",
		},
		{
			name:      "allowed name characters",
			value:     "{user.name-1_x}",
			variables: map[string]string{"user.name-1_x": "ok"},
			want:      "ok",
		},
		{
			name:      "names with spaces are not placeholders",
			value:     "{user name}",
			variables: map[string]string{"user name": "ok"},
			want:      "{user name}",
		},
		{
			name:      "empty value replaces placeholder",
			value:     "[{a}]",
			variables: map[string]string{"a": ""},
			want:      "[]",
		},
		{
			name:      "replacement is not substituted again",
			value:     "{a}",
			variables: map[string]string{"a": "{b}", "b": "x"},
			want:      "{b}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Substitute(tt.value, tt.variables); got != tt.want {
				t.Errorf("Substitute(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestSubstituteRequestReplacesFields(t *testing.T) {
	cfg := Config{}
	cfg.Global.Variable = map[string]string{
		"base":   "https://api.example.com",
		"id":     "42",
		"token":  "abc",
		"hdr":    "Authorization",
		"accept": "application/json",
		"user":   "alice",
		"pass":   "secret",
		"field":  "title",
		"file":   "a.txt",
		"path":   "/tmp/a.txt",
		"type":   "string",
	}
	cfg.Global.Headers = map[string]string{"Accept": "{accept}"}
	cfg.Global.BasicAuth = &BasicAuthConfig{Username: "{user}", Password: "{pass}"}

	spec := RequestConfig{
		URL:       "{base}/users/{id}",
		Query:     map[string]string{"token": "{token}"},
		Headers:   map[string]string{"{hdr}": "Bearer {token}"},
		BasicAuth: &BasicAuthConfig{Username: "{user}", Password: "{pass}"},
		JSON:      json.RawMessage(`{"id":"{id}"}`),
		Form: []FormValue{
			{Name: "{field}", Value: "{id}"},
		},
		FormMultipart: []MultipartFormValue{
			{Name: "{field}", Type: "{type}", Value: "{id}", FileName: "{file}", FilePath: "{path}"},
		},
		MultipartForm: []MultipartFormValue{
			{Name: "{field}", Type: "file", Value: "{id}", FileName: "{file}", FilePath: "{path}"},
		},
	}

	got, gotCfg, err := SubstituteRequest(spec, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if got.URL != "https://api.example.com/users/42" {
		t.Errorf("URL = %q", got.URL)
	}
	if got.Query["token"] != "abc" {
		t.Errorf("Query[token] = %q", got.Query["token"])
	}
	if got.Headers["Authorization"] != "Bearer abc" {
		t.Errorf("Headers = %v", got.Headers)
	}
	if got.BasicAuth == nil || got.BasicAuth.Username != "alice" || got.BasicAuth.Password != "secret" {
		t.Errorf("BasicAuth = %+v", got.BasicAuth)
	}
	if string(got.JSON) != `{"id":"42"}` {
		t.Errorf("JSON = %s", got.JSON)
	}
	if len(got.Form) != 1 || got.Form[0].Name != "title" || got.Form[0].Value != "42" {
		t.Errorf("Form = %+v", got.Form)
	}
	if len(got.FormMultipart) != 1 {
		t.Fatalf("FormMultipart length = %d", len(got.FormMultipart))
	}
	fm := got.FormMultipart[0]
	if fm.Name != "title" || fm.Type != "{type}" || fm.Value != "42" || fm.FileName != "a.txt" || fm.FilePath != "/tmp/a.txt" {
		t.Errorf("FormMultipart = %+v", fm)
	}
	if len(got.MultipartForm) != 1 {
		t.Fatalf("MultipartForm length = %d", len(got.MultipartForm))
	}
	mf := got.MultipartForm[0]
	if mf.Name != "title" || mf.Type != "file" || mf.Value != "42" || mf.FileName != "a.txt" || mf.FilePath != "/tmp/a.txt" {
		t.Errorf("MultipartForm = %+v", mf)
	}

	if gotCfg.Global.Headers["Accept"] != "application/json" {
		t.Errorf("Global.Headers = %v", gotCfg.Global.Headers)
	}
	if gotCfg.Global.BasicAuth == nil || gotCfg.Global.BasicAuth.Username != "alice" || gotCfg.Global.BasicAuth.Password != "secret" {
		t.Errorf("Global.BasicAuth = %+v", gotCfg.Global.BasicAuth)
	}
}

func TestSubstituteRequestRawBodyString(t *testing.T) {
	cfg := Config{}
	cfg.Global.Variable = map[string]string{"name": "alice"}

	spec := RequestConfig{Body: json.RawMessage(`"hello {name}"`)}

	got, _, err := SubstituteRequest(spec, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Body) != `"hello alice"` {
		t.Errorf("Body = %s, want %s", got.Body, `"hello alice"`)
	}

	var body string
	if err := json.Unmarshal(got.Body, &body); err != nil {
		t.Fatalf("Body is not a JSON string: %v", err)
	}
	if body != "hello alice" {
		t.Errorf("unwrapped Body = %q", body)
	}
}

func TestSubstituteRequestRawBodyNonString(t *testing.T) {
	cfg := Config{}
	cfg.Global.Variable = map[string]string{"id": "42"}

	spec := RequestConfig{Body: json.RawMessage(`{"id":{id}}`)}

	got, _, err := SubstituteRequest(spec, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Body) != `{"id":42}` {
		t.Errorf("Body = %s", got.Body)
	}
}

func TestSubstituteRequestKeepsNilFields(t *testing.T) {
	spec := RequestConfig{URL: "http://example.com"}

	got, gotCfg, err := SubstituteRequest(spec, Config{})
	if err != nil {
		t.Fatal(err)
	}
	if got.Headers != nil || got.Query != nil || got.BasicAuth != nil {
		t.Errorf("expected nil headers, query, and basic auth, got %v, %v, %+v", got.Headers, got.Query, got.BasicAuth)
	}
	if got.Form != nil || got.FormMultipart != nil || got.MultipartForm != nil {
		t.Errorf("expected nil form fields, got %v, %v, %v", got.Form, got.FormMultipart, got.MultipartForm)
	}
	if gotCfg.Global.Headers != nil || gotCfg.Global.BasicAuth != nil {
		t.Errorf("expected nil global headers and basic auth, got %v, %+v", gotCfg.Global.Headers, gotCfg.Global.BasicAuth)
	}
}

func TestSubstituteRequestDoesNotMutateInput(t *testing.T) {
	cfg := Config{}
	cfg.Global.Variable = map[string]string{"token": "abc"}
	cfg.Global.Headers = map[string]string{"Authorization": "{token}"}

	spec := RequestConfig{
		URL:     "http://example.com/{token}",
		Headers: map[string]string{"X-Token": "{token}"},
	}

	if _, _, err := SubstituteRequest(spec, cfg); err != nil {
		t.Fatal(err)
	}

	if spec.URL != "http://example.com/{token}" {
		t.Errorf("input URL changed to %q", spec.URL)
	}
	if spec.Headers["X-Token"] != "{token}" {
		t.Errorf("input Headers changed to %v", spec.Headers)
	}
	if cfg.Global.Headers["Authorization"] != "{token}" {
		t.Errorf("input Global.Headers changed to %v", cfg.Global.Headers)
	}
}
