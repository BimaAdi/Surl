package core

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type FormValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func parseFormBody(req *http.Request, values []FormValue) {
	if len(values) == 0 {
		return
	}
	// Encode in declaration order; url.Values.Encode would sort the keys.
	var encoded strings.Builder
	for _, value := range values {
		if encoded.Len() > 0 {
			encoded.WriteByte('&')
		}
		encoded.WriteString(url.QueryEscape(value.Name))
		encoded.WriteByte('=')
		encoded.WriteString(url.QueryEscape(value.Value))
	}
	body := []byte(encoded.String())
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
}
