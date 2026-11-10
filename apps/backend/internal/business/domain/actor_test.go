package domain

import (
	"errors"
	"testing"
)

func TestRoleIsValid(t *testing.T) {
	for _, role := range []Role{RoleAdmin, RoleManager, RoleCashier} {
		if !role.IsValid() {
			t.Errorf("%q.IsValid() = false, want true", role)
		}
	}
	for _, role := range []Role{"", "root", "Admin"} {
		if role.IsValid() {
			t.Errorf("%q.IsValid() = true, want false", role)
		}
	}
}

func TestActorAuthorize(t *testing.T) {
	cashier := Actor{ID: "u1", Name: "Ana", Role: RoleCashier}

	if err := cashier.Authorize(RoleAdmin, RoleCashier); err != nil {
		t.Fatalf("Authorize with a held role = %v, want nil", err)
	}
	if err := cashier.Authorize(RoleAdmin, RoleManager); !errors.Is(err, ErrForbidden) {
		t.Fatalf("Authorize without the role = %v, want ErrForbidden", err)
	}
	if err := cashier.Authorize(); !errors.Is(err, ErrForbidden) {
		t.Fatalf("Authorize with no allowed roles = %v, want ErrForbidden", err)
	}

	anonymous := Actor{Role: RoleAdmin}
	if err := anonymous.Authorize(RoleAdmin); !errors.Is(err, ErrForbidden) {
		t.Fatalf("Authorize for an actor without ID = %v, want ErrForbidden", err)
	}
}
