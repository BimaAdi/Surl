package core

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func parseRawBody(req *http.Request, value json.RawMessage, variables map[string]string) {
	value = bytes.TrimSpace(value)
	if len(value) == 0 || string(value) == "null" {
		return
	}

	// The body field is a JSON string in the configuration, so unwrap it
	// before applying variable substitution and sending it as raw content.
	var raw string
	if err := json.Unmarshal(value, &raw); err == nil {
		value = []byte(raw)
	}
	body := []byte(Substitute(string(value), variables))
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
}
