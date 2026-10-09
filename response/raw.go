package response

import (
	"fmt"
	"sort"
	"strings"

	"github.com/BimaAdi/surl/core"
)

func RawResponse(resp core.Response, verbose bool) ([]byte, error) {
	if !verbose {
		return []byte(resp.Body), nil
	}

	keys := make([]string, 0, len(resp.Header))
	for key := range resp.Header {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var b strings.Builder
	fmt.Fprintf(&b, "status: %d\n", resp.Status)
	for _, key := range keys {
		fmt.Fprintf(&b, "%s: %s\n", key, resp.Header[key])
	}
	b.WriteString("\n")
	b.WriteString(resp.Body)
	return []byte(b.String()), nil
}
