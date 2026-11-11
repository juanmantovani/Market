package domain

import "slices"

// Permission is something an actor may do. Use cases ask for a permission,
// never for a role, so which role may do what is decided only here.
type Permission string

const (
	PermissionViewCatalog   Permission = "catalog:view"
	PermissionManageCatalog Permission = "catalog:manage"

	PermissionViewCustomers   Permission = "customers:view"
	PermissionCreateCustomers Permission = "customers:create"
	PermissionManageCustomers Permission = "customers:manage"
)

// rolePermissions is the only place that grants permissions. Anyone who
// sells reads the catalog and registers customers at the counter; only
// managers maintain the catalog and edit or delete customers.
var rolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionViewCatalog, PermissionManageCatalog,
		PermissionViewCustomers, PermissionCreateCustomers, PermissionManageCustomers,
	},
	RoleManager: {
		PermissionViewCatalog, PermissionManageCatalog,
		PermissionViewCustomers, PermissionCreateCustomers, PermissionManageCustomers,
	},
	RoleCashier: {
		PermissionViewCatalog,
		PermissionViewCustomers, PermissionCreateCustomers,
	},
}

func (role Role) Can(permission Permission) bool {
	return slices.Contains(rolePermissions[role], permission)
}
