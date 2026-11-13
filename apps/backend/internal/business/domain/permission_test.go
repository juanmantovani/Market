package domain

import "testing"

// TestRolePermissions pins the whole matrix: changing who may do what must
// show up as a change here.
func TestRolePermissions(t *testing.T) {
	permissions := []Permission{
		PermissionViewCatalog, PermissionManageCatalog,
		PermissionViewCustomers, PermissionCreateCustomers, PermissionManageCustomers,
		PermissionViewSuppliers, PermissionManageSuppliers,
		PermissionViewSellers, PermissionManageSellers,
		PermissionViewStock, PermissionManageStock,
		PermissionViewSales, PermissionCreateSales, PermissionVoidSales,
	}
	want := map[Role][]Permission{
		RoleAdmin:   permissions,
		RoleManager: permissions,
		RoleCashier: {
			PermissionViewCatalog, PermissionViewCustomers, PermissionCreateCustomers, PermissionViewSellers, PermissionViewStock,
			PermissionViewSales, PermissionCreateSales,
		},
		"root": {},
	}

	for role, granted := range want {
		for _, permission := range permissions {
			expected := false
			for _, grant := range granted {
				expected = expected || grant == permission
			}
			if got := role.Can(permission); got != expected {
				t.Errorf("%q.Can(%q) = %v, want %v", role, permission, got, expected)
			}
		}
	}
}
