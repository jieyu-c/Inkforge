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

func PromptCreated(tenantID, nsSlug, promptKey string) {
	logx.Infof("audit event=prompt.created tenant_id=%s ns_slug=%s prompt_key=%s", tenantID, nsSlug, promptKey)
}

func PromptDeleted(tenantID, nsSlug, promptKey string) {
	logx.Infof("audit event=prompt.deleted tenant_id=%s ns_slug=%s prompt_key=%s", tenantID, nsSlug, promptKey)
}

func PromptDraftSaved(tenantID, nsSlug, promptKey string) {
	logx.Infof("audit event=prompt.updated tenant_id=%s ns_slug=%s prompt_key=%s detail=draft", tenantID, nsSlug, promptKey)
}

func PromptVersionCreated(tenantID, nsSlug, promptKey, versionLabel string) {
	logx.Infof("audit event=prompt.version.created tenant_id=%s ns_slug=%s prompt_key=%s version=%s",
		tenantID, nsSlug, promptKey, versionLabel)
}

func PromptPointerChanged(tenantID, nsSlug, promptKey, channel, fromVer, toVer string) {
	logx.Infof("audit event=prompt.pointer.changed tenant_id=%s ns_slug=%s prompt_key=%s channel=%s from=%s to=%s",
		tenantID, nsSlug, promptKey, channel, fromVer, toVer)
}
