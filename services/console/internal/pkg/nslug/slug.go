package nslug

import (
	"errors"
	"regexp"
	"strings"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
)

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
