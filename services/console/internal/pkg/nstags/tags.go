package nstags

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
)

func ToNull(tags []string) (sql.NullString, error) {
	if len(tags) == 0 {
		return sql.NullString{}, nil
	}
	for _, t := range tags {
		ts := strings.TrimSpace(t)
		if ts == "" {
			continue
		}
		if len(ts) > 128 {
			return sql.NullString{}, apperr.BadRequest("INVALID_NS_TAGS", "Each tag must be ≤128 chars")
		}
	}
	b, err := json.Marshal(tags)
	if err != nil {
		return sql.NullString{}, apperr.BadRequest("INVALID_NS_TAGS", "Tags must serialize as JSON")
	}
	return sql.NullString{String: string(b), Valid: true}, nil
}

func FromNull(tags sql.NullString) ([]string, error) {
	if !tags.Valid || strings.TrimSpace(tags.String) == "" {
		return nil, nil
	}
	var out []string
	if err := json.Unmarshal([]byte(tags.String), &out); err != nil {
		return nil, nil
	}
	return out, nil
}
