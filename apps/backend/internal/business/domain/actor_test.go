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

	if err := cashier.Authorize(PermissionViewCatalog); err != nil {
		t.Fatalf("Authorize(granted) = %v, want nil", err)
	}
	if err := cashier.Authorize(PermissionManageCatalog); !errors.Is(err, ErrForbidden) {
		t.Fatalf("Authorize(not granted) = %v, want ErrForbidden", err)
	}

	anonymous := Actor{Role: RoleAdmin}
	if err := anonymous.Authorize(PermissionViewCatalog); !errors.Is(err, ErrForbidden) {
		t.Fatalf("Authorize for an actor without ID = %v, want ErrForbidden", err)
	}

	unknownRole := Actor{ID: "u2", Role: "root"}
	if err := unknownRole.Authorize(PermissionViewCatalog); !errors.Is(err, ErrForbidden) {
		t.Fatalf("Authorize for an unknown role = %v, want ErrForbidden", err)
	}
}
