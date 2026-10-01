package core

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type MultipartFormValue struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"value"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
}

func parseMultipartBody(req *http.Request, values []MultipartFormValue, variables map[string]string) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for _, value := range values {
		name := Substitute(value.Name, variables)
		switch value.Type {
		case "string":
			if err := writer.WriteField(name, Substitute(value.Value, variables)); err != nil {
				return fmt.Errorf("write multipart field %q: %w", name, err)
			}
		case "file":
			fileName := Substitute(value.FileName, variables)
			filePath := Substitute(value.FilePath, variables)
			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("open multipart file %q: %w", filePath, err)
			}
			part, createErr := writer.CreateFormFile(name, fileName)
			if createErr != nil {
				_ = file.Close()
				return fmt.Errorf("create multipart file field %q: %w", name, createErr)
			}
			_, copyErr := io.Copy(part, file)
			closeErr := file.Close()
			if copyErr != nil {
				return fmt.Errorf("copy multipart file %q: %w", filePath, copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("close multipart file %q: %w", filePath, closeErr)
			}
		default:
			return fmt.Errorf("unsupported multipart form type %q", value.Type)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("finish multipart form: %w", err)
	}
	encodedBody := body.Bytes()
	req.Body = io.NopCloser(bytes.NewReader(encodedBody))
	req.ContentLength = int64(len(encodedBody))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(encodedBody)), nil
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return nil
}
