package nsaudit

import "github.com/zeromicro/go-zero/core/logx"

func Created(tenantID, nsSlug string) {
	logx.Infof("audit event=ns.created tenant_id=%s ns_slug=%s", tenantID, nsSlug)
}

func Archived(tenantID, nsSlug string) {
	logx.Infof("audit event=ns.archived tenant_id=%s ns_slug=%s", tenantID, nsSlug)
}

func Restored(tenantID, nsSlug string) {
	logx.Infof("audit event=ns.restored tenant_id=%s ns_slug=%s", tenantID, nsSlug)
}

func SettingsUpdated(tenantID, nsSlug string) {
	logx.Infof("audit event=ns.settings.updated tenant_id=%s ns_slug=%s", tenantID, nsSlug)
}

func PromptQuotaExceeded(tenantID, nsSlug string) {
	logx.Infof("audit event=quota.exceeded subtype=prompt tenant_id=%s ns_slug=%s", tenantID, nsSlug)
}
