package response

import (
	"bytes"
	"encoding/json"
)

// PrettifyJSON indents data when it is valid JSON. Data that is not valid JSON
// is returned unchanged.
func PrettifyJSON(data []byte) []byte {
	if !json.Valid(data) {
		return data
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		return data
	}
	return buf.Bytes()
}
