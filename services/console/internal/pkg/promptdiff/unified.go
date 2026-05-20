package promptdiff

import (
	"bytes"
	"fmt"
	"strings"
)

// Unified returns a readable line-oriented diff; empty string if equal.
func Unified(labelA, labelB, a, b string) string {
	if a == b {
		return ""
	}
	la := strings.Split(a, "\n")
	lb := strings.Split(b, "\n")
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "--- %s\n+++ %s\n", labelA, labelB)
	max := len(la)
	if len(lb) > max {
		max = len(lb)
	}
	for i := 0; i < max; i++ {
		var sa, sb string
		if i < len(la) {
			sa = la[i]
		}
		if i < len(lb) {
			sb = lb[i]
		}
		switch {
		case sa == sb:
			fmt.Fprintf(&buf, " %s\n", sa)
		case i >= len(la):
			fmt.Fprintf(&buf, "+ %s\n", sb)
		case i >= len(lb):
			fmt.Fprintf(&buf, "- %s\n", sa)
		default:
			fmt.Fprintf(&buf, "- %s\n+ %s\n", sa, sb)
		}
	}
	return buf.String()
}
