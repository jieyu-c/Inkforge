package promptdraft

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
)

const (
	ModeStrict = "strict"
	ModeWarn   = "warn"
)

var placeholderRe = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

type Result struct {
	Warnings []string
}

type varEntry struct {
	Name string `json:"name"`
}

// Validate checks placeholders in body against JSON schema (array of { "name": "..." }).
func Validate(body, schemaJSON, mode string) (*Result, error) {
	schemaJSON = strings.TrimSpace(schemaJSON)
	if schemaJSON == "" {
		schemaJSON = "[]"
	}
	var entries []varEntry
	if err := json.Unmarshal([]byte(schemaJSON), &entries); err != nil {
		return nil, apperr.BadRequest("INVALID_PROMPT_SCHEMA", "draft schema must be a JSON array of variable objects")
	}
	seen := map[string]struct{}{}
	declared := map[string]struct{}{}
	for _, e := range entries {
		n := strings.TrimSpace(e.Name)
		if n == "" {
			return nil, apperr.BadRequest("INVALID_PROMPT_SCHEMA", "each schema entry needs a non-empty name")
		}
		if _, dup := seen[n]; dup {
			return nil, apperr.BadRequest("INVALID_PROMPT_SCHEMA", "duplicate variable name in schema: "+n)
		}
		seen[n] = struct{}{}
		declared[n] = struct{}{}
	}
	used := map[string]struct{}{}
	for _, m := range placeholderRe.FindAllStringSubmatch(body, -1) {
		if len(m) > 1 {
			used[m[1]] = struct{}{}
		}
	}
	var undeclared, unused []string
	for u := range used {
		if _, ok := declared[u]; !ok {
			undeclared = append(undeclared, u)
		}
	}
	for d := range declared {
		if _, ok := used[d]; !ok {
			unused = append(unused, d)
		}
	}
	res := &Result{}
	if len(undeclared) > 0 {
		msg := "placeholders in body without schema entry: " + strings.Join(undeclared, ", ")
		if mode == ModeWarn {
			res.Warnings = append(res.Warnings, msg)
		} else {
			return nil, apperr.BadRequest("PROMPT_PLACEHOLDER_MISMATCH", msg)
		}
	}
	if len(unused) > 0 {
		msg := "schema variables not used in body: " + strings.Join(unused, ", ")
		if mode == ModeWarn {
			res.Warnings = append(res.Warnings, msg)
		} else {
			return nil, apperr.BadRequest("PROMPT_UNUSED_SCHEMA_VARS", msg)
		}
	}
	return res, nil
}
