package response

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/BimaAdi/surl/core"
)

type verboseResponse struct {
	Status int               `json:"status"`
	Header map[string]string `json:"header"`
	Body   any               `json:"body"`
}

func ResponseJSON(res core.Response, verbose bool) ([]byte, error) {
	var v any
	if verbose {
		v = verboseResponse{Status: res.Status, Header: res.Header, Body: bodyValue(res)}
	} else {
		v = bodyValue(res)
	}
	encoded, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode response: %w", err)
	}
	return encoded, nil
}

func bodyValue(resp core.Response) any {
	if isJSONContentType(headerValue(resp.Header, "Content-Type")) && json.Valid([]byte(resp.Body)) {
		return json.RawMessage(resp.Body)
	}
	return resp.Body
}

func headerValue(header map[string]string, name string) string {
	for key, value := range header {
		if strings.EqualFold(key, name) {
			return value
		}
	}
	return ""
}

func isJSONContentType(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "application/json")
}
