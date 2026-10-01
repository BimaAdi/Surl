package core

import (
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestExecuteRequestUsesMultipartFormBody(t *testing.T) {
	filePath := t.TempDir() + "/upload.txt"
	if err := os.WriteFile(filePath, []byte("file contents"), 0600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "multipart/form-data" || params["boundary"] == "" {
			t.Errorf("Content-Type = %q, error = %v", r.Header.Get("Content-Type"), err)
			return
		}

		reader := multipart.NewReader(r.Body, params["boundary"])
		field, err := reader.NextPart()
		if err != nil {
			t.Errorf("read string part: %v", err)
			return
		}
		fieldValue, _ := io.ReadAll(field)
		if field.FormName() != "username" || string(fieldValue) != "hello" {
			t.Errorf("string part = (%q, %q)", field.FormName(), fieldValue)
		}

		file, err := reader.NextPart()
		if err != nil {
			t.Errorf("read file part: %v", err)
			return
		}
		fileContents, _ := io.ReadAll(file)
		if file.FormName() != "upload" || file.FileName() != "renamed.txt" || string(fileContents) != "file contents" {
			t.Errorf("file part = (%q, %q, %q)", file.FormName(), file.FileName(), fileContents)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	_, err := ExecuteRequest(context.Background(), server.Client(), RequestConfig{
		URL:    server.URL,
		Method: http.MethodPost,
		FormMultipart: []MultipartFormValue{
			{Name: "username", Type: "string", Value: "hello"},
			{Name: "upload", Type: "file", FileName: "renamed.txt", FilePath: filePath},
		},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}
