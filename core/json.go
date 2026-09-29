package core

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func parseJSONBody(req *http.Request, value json.RawMessage, variables map[string]string) {
	value = bytes.TrimSpace(value)
	if len(value) == 0 || string(value) == "null" {
		return
	}
	body := []byte(Substitute(string(value), variables))
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
}
