package promptversion

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
)

// Semver core: major.minor.patch with optional pre-release suffix.
var semverCore = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z][0-9A-Za-z.-]*))?$`)

// NormalizeLabel strips whitespace and optional leading "v".
func NormalizeLabel(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(strings.TrimPrefix(s, "V"), "v")
	return s
}

// ValidateLabel ensures major.minor.patch (e.g. 1.2.3).
func ValidateLabel(label string) error {
	if label == "" {
		return apperr.BadRequest("INVALID_VERSION", "version is required")
	}
	if len(label) > 64 {
		return apperr.BadRequest("INVALID_VERSION", "version must be at most 64 characters")
	}
	if !semverCore.MatchString(label) {
		return apperr.BadRequest("INVALID_VERSION", "version must be semver format major.minor.patch (e.g. 1.2.3)")
	}
	return nil
}

type triple [3]int

func parseTriple(label string) (triple, bool) {
	m := semverCore.FindStringSubmatch(label)
	if m == nil {
		return triple{}, false
	}
	a, _ := strconv.Atoi(m[1])
	b, _ := strconv.Atoi(m[2])
	c, _ := strconv.Atoi(m[3])
	return triple{a, b, c}, true
}

func (a triple) less(b triple) bool {
	if a[0] != b[0] {
		return a[0] < b[0]
	}
	if a[1] != b[1] {
		return a[1] < b[1]
	}
	return a[2] < b[2]
}

func (t triple) String() string {
	return fmt.Sprintf("%d.%d.%d", t[0], t[1], t[2])
}

func maxTriple(existing []string) (triple, bool) {
	var best triple
	var ok bool
	for _, raw := range existing {
		l := NormalizeLabel(raw)
		if l == "" {
			continue
		}
		t, parsed := parseTriple(l)
		if !parsed {
			continue
		}
		if !ok || best.less(t) {
			best = t
			ok = true
		}
	}
	return best, ok
}

// AutoNextLabel returns the next default label: 1.0.0, or patch+1 of the greatest semver.
func AutoNextLabel(existing []string) string {
	if best, ok := maxTriple(existing); ok {
		best[2]++
		return best.String()
	}
	return "1.0.0"
}

// ResolveLabel picks explicit semver or AutoNextLabel when explicit is empty.
func ResolveLabel(explicit string, existing []string) (string, error) {
	raw := NormalizeLabel(explicit)
	if raw == "" {
		return AutoNextLabel(existing), nil
	}
	if err := ValidateLabel(raw); err != nil {
		return "", err
	}
	for _, e := range existing {
		if NormalizeLabel(e) == raw {
			return "", apperr.Conflict("VERSION_LABEL_TAKEN", "version already exists for this prompt")
		}
	}
	return raw, nil
}
