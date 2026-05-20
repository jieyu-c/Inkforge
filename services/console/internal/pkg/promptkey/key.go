package promptkey

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
)

// Key syntax: lowercase labels with digits, underscore, hyphen; DNS-ish; 3–128 chars.
var keySyntax = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9_-]*[a-z0-9])?$`)

func Normalize(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// FromLabel derives a prompt_key candidate from a human title (may need uniquify).
func FromLabel(label string) string {
	s := strings.ToLower(strings.TrimSpace(label))
	s = nonAlnum.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	if len(s) < 3 {
		return "prompt-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:8]
	}
	if len(s) > 128 {
		s = strings.Trim(s[:128], "_")
	}
	if !keySyntax.MatchString(s) {
		return "prompt-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:8]
	}
	return s
}

func candidate(base string, n int) string {
	if n <= 0 {
		return base
	}
	suffix := fmt.Sprintf("-%d", n+1)
	max := 128 - len(suffix)
	if max < 3 {
		return base
	}
	b := base
	if len(b) > max {
		b = strings.Trim(b[:max], "_-")
	}
	return b + suffix
}

// AllocateUnique returns the first available key under exists(key)==true when taken.
func AllocateUnique(label string, exists func(string) bool) (string, error) {
	base := FromLabel(label)
	for i := 0; i < 32; i++ {
		c := candidate(base, i)
		valid, err := MustValidate(c)
		if err != nil {
			return "", err
		}
		if !exists(valid) {
			return valid, nil
		}
	}
	return "", apperr.Conflict("PROMPT_KEY_TAKEN", "could not allocate a unique prompt_key")
}

func MustValidate(raw string) (string, error) {
	s := Normalize(raw)
	if s == "" {
		return "", apperr.BadRequest("INVALID_PROMPT_KEY", "prompt_key is required")
	}
	switch {
	case len(s) < 3:
		return "", apperr.BadRequest("INVALID_PROMPT_KEY", "prompt_key must be at least 3 characters")
	case len(s) > 128:
		return "", apperr.BadRequest("INVALID_PROMPT_KEY", "prompt_key must be at most 128 characters")
	case !keySyntax.MatchString(s):
		return "", apperr.BadRequest("INVALID_PROMPT_KEY",
			"prompt_key must be lowercase letters, digits, hyphen, and underscore")
	default:
		return s, nil
	}
}
