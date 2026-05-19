package nsquota

import (
	"database/sql"
	"testing"
)

func TestEffectivePromptsMax(t *testing.T) {
	def := int64(100)
	plat := int64(50)
	got := EffectivePromptsMax(def, plat, sql.NullInt64{})
	if got != 50 {
		t.Fatal(got)
	}
	row := sql.NullInt64{Valid: true, Int64: 30}
	got2 := EffectivePromptsMax(def, plat, row)
	if got2 != 30 {
		t.Fatal(got2)
	}
	got3 := EffectivePromptsMax(0, 0, sql.NullInt64{})
	if got3 != 0 {
		t.Fatal(got3)
	}
}

func TestPromptCapacityExceeded(t *testing.T) {
	if !PromptCapacityExceeded(50, 50) {
		t.Fatal()
	}
	if PromptCapacityExceeded(49, 50) {
		t.Fatal()
	}
	if PromptCapacityExceeded(999, 0) {
		t.Fatal()
	}
}
