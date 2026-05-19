package nsquota

import (
	"database/sql"
)

// EffectivePromptsMax merges NS row config, defaults, and optional platform-wide cap (from config YAML).
func EffectivePromptsMax(defaultPerTenant int64, platformCap int64, rowConfigured sql.NullInt64) int64 {
	var effective int64
	switch {
	case rowConfigured.Valid && rowConfigured.Int64 > 0:
		effective = rowConfigured.Int64
	case defaultPerTenant > 0:
		effective = defaultPerTenant
	default:
		effective = 0
	}
	if platformCap > 0 && (effective == 0 || effective > platformCap) {
		return platformCap
	}
	return effective
}

func PromptCapacityExceeded(promptCount uint64, effectiveMax int64) bool {
	if effectiveMax <= 0 {
		return false
	}
	return int64(promptCount) >= effectiveMax
}
