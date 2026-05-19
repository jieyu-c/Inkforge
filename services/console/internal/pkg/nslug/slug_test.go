package nslug

import (
	"testing"
)

func TestMustValidate_Good(t *testing.T) {
	s, err := MustValidate("  My-Service_Prod ")
	if err != nil {
		t.Fatal(err)
	}
	if s != "my-service-prod" {
		t.Fatal(s)
	}
}

func TestMustValidate_Invalid(t *testing.T) {
	_, err := MustValidate("$bad")
	if err == nil {
		t.Fatal("expected error")
	}
}
