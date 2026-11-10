package domain

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleManager Role = "manager"
	RoleCashier Role = "cashier"
)

func (role Role) IsValid() bool {
	switch role {
	case RoleAdmin, RoleManager, RoleCashier:
		return true
	default:
		return false
	}
}

// Actor is whoever performs an operation. Every use case receives one so
// permissions and audit fields don't depend on how the caller authenticated.
type Actor struct {
	ID   string
	Name string
	Role Role
}

func (actor Actor) HasRole(roles ...Role) bool {
	for _, role := range roles {
		if actor.Role == role {
			return true
		}
	}

	return false
}

// Authorize returns ErrForbidden unless the actor holds one of roles.
func (actor Actor) Authorize(roles ...Role) error {
	if actor.ID == "" || !actor.HasRole(roles...) {
		return ErrForbidden
	}

	return nil
}
