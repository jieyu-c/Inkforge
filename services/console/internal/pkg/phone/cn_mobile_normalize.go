package phone

import (
	"regexp"
	"strings"
)

var cnTail = regexp.MustCompile(`^1[3-9]\d{9}$`)

func digits(raw string) string {
	var b strings.Builder
	for _, r := range strings.TrimSpace(raw) {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Canonical returns normalized key +861XXXXXXXXXX for mainland mobiles, empty if invalid.
func Canonical(raw string) string {
	d := digits(raw)
	switch len(d) {
	case 11:
		if cnTail.MatchString(d) {
			d = "86" + d
			break
		}
		return ""
	case 13:
		if !strings.HasPrefix(d, "86") {
			return ""
		}
		tail := d[2:]
		if !cnTail.MatchString(tail) {
			return ""
		}
	default:
		return ""
	}

	return "+" + d
}
