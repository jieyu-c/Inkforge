package namespace

import (
	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsquota"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nstags"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"
)

func detailFromRow(svcCtx *svc.ServiceContext, row *model.ConsoleNamespaces) (*types.NamespaceDetail, error) {
	cfg := svcCtx.Config.Namespace
	effective := nsquota.EffectivePromptsMax(
		cfg.DefaultQuotaPromptsPerNs,
		cfg.PlatformQuotaPromptsCap,
		row.QuotaPromptsMax,
	)
	tags, err := nstags.FromNull(row.Tags)
	if err != nil {
		return nil, err
	}
	d := types.NamespaceDetail{
		NsSlug:          row.NsSlug,
		DisplayName:     row.DisplayName,
		Status:          row.Status,
		QuotaPromptsMax: effective,
		PromptCount:     int64(row.PromptCount),
		Tags:            tags,
	}
	if row.Description.Valid {
		d.Description = row.Description.String
	}
	if row.DefaultChannelSlug.Valid {
		d.DefaultChannelSlug = row.DefaultChannelSlug.String
	}
	if row.ArchivedAt.Valid {
		d.ArchivedAtIso = row.ArchivedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return &d, nil
}
