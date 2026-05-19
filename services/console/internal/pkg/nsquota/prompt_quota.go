package nsquota

import (
	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
)

// PromptCreateRejected returns an error suitable for Create Prompt handlers (NS-004 P0 hooks).
func PromptCreateRejected(
	defaultPerTenant int64, platformCap int64,
	row *model.ConsoleNamespaces,
) error {
	if row == nil {
		return apperr.BadRequest("NS_REQUIRED", "Namespace is required")
	}
	cap := EffectivePromptsMax(defaultPerTenant, platformCap, row.QuotaPromptsMax)
	if PromptCapacityExceeded(row.PromptCount, cap) {
		return apperr.Conflict("NS_QUOTA_PROMPTS_EXCEEDED", "Prompt count quota exceeded for this namespace")
	}
	return nil
}
