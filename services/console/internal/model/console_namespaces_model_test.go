package model

import (
	"strings"
	"testing"
)

func TestListNamespacesSqlShapeContainsTenant(t *testing.T) {
	s := ListNamespacesSqlShape()
	if !strings.Contains(s, "tenant_id = ?") {
		t.Fatal(s)
	}
}
