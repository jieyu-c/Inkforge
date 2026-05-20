package promptversion

import "testing"

func TestResolveLabel_Auto(t *testing.T) {
	got, err := ResolveLabel("", []string{"1.0.0", "1.2.3"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.2.4" {
		t.Fatalf("got %q want 1.2.4", got)
	}
}

func TestResolveLabel_Custom(t *testing.T) {
	got, err := ResolveLabel("v2.0.0", []string{"1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "2.0.0" {
		t.Fatalf("got %q", got)
	}
}

func TestValidateLabel_Invalid(t *testing.T) {
	if err := ValidateLabel("1.2"); err == nil {
		t.Fatal("expected error for 1.2")
	}
}
