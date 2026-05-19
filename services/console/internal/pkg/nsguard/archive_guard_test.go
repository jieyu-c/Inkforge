package nsguard

import (
	"errors"
	"testing"

	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
)

func TestRejectArchivedWrite(t *testing.T) {
	err := RejectArchivedStatus(StatusArchived, "writing prompts")
	if err == nil {
		t.Fatal()
	}
	var he *apperr.HTTP
	if !errors.As(err, &he) || he.Code != "NS_ARCHIVED_WRITE_REJECTED" {
		t.Fatal(err)
	}
}
