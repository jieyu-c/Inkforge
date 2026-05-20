package nslug

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
)

// DefaultChannelSlug is assigned to new namespaces when the client does not manage channels.
const DefaultChannelSlug = "production"

// DNS-label style: lowercase alnum/hyphens, no edge hyphens (enforced loosely), length checked separately.
var slugSyntax = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)

func Normalize(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, "_", "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	return s
}

func Validate(norm string) error {
	switch {
	case len(norm) < 3:
		return apperr.BadRequest("INVALID_NS_SLUG", "Namespace slug must be at least 3 characters")
	case len(norm) > 63:
		return apperr.BadRequest("INVALID_NS_SLUG", "Namespace slug must be at most 63 characters")
	case slugSyntax.MatchString(norm):
		return nil
	default:
		return apperr.BadRequest("INVALID_NS_SLUG",
			"Namespace slug must be lowercase letters, digits, and hyphen (DNS-label style)")
	}
}

// FromDisplayName derives an ns_slug candidate from a display name (may need uniquify).
func FromDisplayName(displayName string) string {
	n := Normalize(displayName)
	if len(n) >= 3 && Validate(n) == nil {
		return n
	}
	return fmt.Sprintf("ns-%s", strings.ReplaceAll(uuid.NewString(), "-", "")[:8])
}

func candidate(base string, n int) string {
	if n <= 0 {
		return base
	}
	suffix := fmt.Sprintf("-%d", n+1)
	max := 63 - len(suffix)
	if max < 3 {
		return base
	}
	b := base
	if len(b) > max {
		b = strings.Trim(b[:max], "-")
	}
	return b + suffix
}

// AllocateUnique returns the first slug not taken according to exists(slug).
func AllocateUnique(displayName string, exists func(string) bool) (string, error) {
	base := FromDisplayName(displayName)
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
	return "", apperr.Conflict("NS_SLUG_TAKEN", "could not allocate a unique namespace slug")
}

func MustValidate(raw string) (string, error) {
	n := Normalize(raw)
	if n == "" {
		return "", apperr.BadRequest("INVALID_NS_SLUG", "Namespace slug is required")
	}
	err := Validate(n)
	if err != nil {
		var ze *apperr.HTTP
		if errors.As(err, &ze) && ze.Code == "INVALID_NS_SLUG" {
			return "", err
		}
		return "", err
	}
	return n, nil
}
