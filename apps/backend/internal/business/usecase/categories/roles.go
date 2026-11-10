package categories

import "market/backend/internal/business/domain"

var (
	// Anyone who sells needs the catalog; only managers maintain it.
	readRoles   = []domain.Role{domain.RoleAdmin, domain.RoleManager, domain.RoleCashier}
	manageRoles = []domain.Role{domain.RoleAdmin, domain.RoleManager}
)
