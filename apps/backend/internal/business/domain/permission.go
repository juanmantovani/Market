package domain

import "slices"

// Permission is something an actor may do. Use cases ask for a permission,
// never for a role, so which role may do what is decided only here.
type Permission string

const (
	PermissionViewCatalog   Permission = "catalog:view"
	PermissionManageCatalog Permission = "catalog:manage"
)

// rolePermissions is the only place that grants permissions. Anyone who
// sells needs to read the catalog; only managers maintain it.
var rolePermissions = map[Role][]Permission{
	RoleAdmin:   {PermissionViewCatalog, PermissionManageCatalog},
	RoleManager: {PermissionViewCatalog, PermissionManageCatalog},
	RoleCashier: {PermissionViewCatalog},
}

func (role Role) Can(permission Permission) bool {
	return slices.Contains(rolePermissions[role], permission)
}
