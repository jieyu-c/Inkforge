package phone

import "strings"

// MaskDisplay masks a Canonical mainland mobile (+861…) for HTTP display: first 3 and last 4
// digits of the 11‑digit subscriber number remain; middle 4 replaced by asterisks.
// Unexpected inputs are heavily masked so the original is not echoed.
func MaskDisplay(canonical string) string {
	d := digits(canonical)
	var tail string
	switch len(d) {
	case 13:
		if strings.HasPrefix(d, "86") && cnTail.MatchString(d[2:]) {
			tail = d[2:]
		}
	case 11:
		if cnTail.MatchString(d) {
			tail = d
		}
	default:
		return maskUnexpected(d)
	}
	if tail == "" {
		return maskUnexpected(d)
	}
	return "+86" + tail[:3] + "****" + tail[len(tail)-4:]
}

func maskUnexpected(d string) string {
	if len(d) == 0 {
		return "***"
	}
	if len(d) <= 4 {
		return strings.Repeat("*", len(d))
	}
	return strings.Repeat("*", len(d)-4) + d[len(d)-4:]
}
