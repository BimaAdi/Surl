// Package tutorial embeds the surl tutorial document so the CLI can print it.
package tutorial

import _ "embed"

//go:embed tutorial.md
var Content string
