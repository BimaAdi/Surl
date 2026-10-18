package request

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

// ParseReadFileBody reads the file at path and sends its content as the
// request body, the same way a raw body is sent.
func ParseReadFileBody(req *http.Request, path string) error {
	if path == "" {
		return nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read body file %q: %w", path, err)
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	return nil
}
