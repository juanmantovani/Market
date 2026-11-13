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

	PermissionViewSuppliers   Permission = "suppliers:view"
	PermissionManageSuppliers Permission = "suppliers:manage"

	PermissionViewSellers   Permission = "sellers:view"
	PermissionManageSellers Permission = "sellers:manage"

	PermissionViewStock   Permission = "stock:view"
	PermissionManageStock Permission = "stock:manage"

	PermissionViewSales   Permission = "sales:view"
	PermissionCreateSales Permission = "sales:create"
	PermissionVoidSales   Permission = "sales:void"

	PermissionViewDashboard Permission = "dashboard:view"
)

// rolePermissions is the only place that grants permissions. Anyone who
// sells reads the catalog, checks the stock, picks the seller, registers
// customers at the counter, looks up past sales and sees the day's summary;
// only managers maintain
// the catalog and the sellers, edit or delete customers, deal with
// suppliers, receive or adjust stock and void sales.
var rolePermissions = map[Role][]Permission{
	RoleAdmin: {
		PermissionViewCatalog, PermissionManageCatalog,
		PermissionViewCustomers, PermissionCreateCustomers, PermissionManageCustomers,
		PermissionViewSuppliers, PermissionManageSuppliers,
		PermissionViewSellers, PermissionManageSellers,
		PermissionViewStock, PermissionManageStock,
		PermissionViewSales, PermissionCreateSales, PermissionVoidSales,
		PermissionViewDashboard,
	},
	RoleManager: {
		PermissionViewCatalog, PermissionManageCatalog,
		PermissionViewCustomers, PermissionCreateCustomers, PermissionManageCustomers,
		PermissionViewSuppliers, PermissionManageSuppliers,
		PermissionViewSellers, PermissionManageSellers,
		PermissionViewStock, PermissionManageStock,
		PermissionViewSales, PermissionCreateSales, PermissionVoidSales,
		PermissionViewDashboard,
	},
	RoleCashier: {
		PermissionViewCatalog,
		PermissionViewCustomers, PermissionCreateCustomers,
		PermissionViewSellers,
		PermissionViewStock,
		PermissionViewSales, PermissionCreateSales,
		PermissionViewDashboard,
	},
}

func (role Role) Can(permission Permission) bool {
	return slices.Contains(rolePermissions[role], permission)
}
