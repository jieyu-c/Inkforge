package nsguard

import (
	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
)

const (
	StatusActive    = "active"
	StatusArchived  = "archived"
	StatusWriteRule = StatusArchived // PRD N-DEC-02 default assumption
)

func RejectArchivedWrite(ns *model.ConsoleNamespaces, opHuman string) error {
	if ns == nil {
		return apperr.NotFound("NS_NOT_FOUND", "Namespace does not exist")
	}
	return RejectArchivedStatus(ns.Status, opHuman)
}

func RejectArchivedStatus(status string, opHuman string) error {
	if status != StatusArchived {
		return nil
	}
	msg := "This namespace is archived; " + opHuman + "."
	return apperr.Conflict("NS_ARCHIVED_WRITE_REJECTED", msg)
}

func IsArchived(status string) bool {
	return status == StatusArchived
}
