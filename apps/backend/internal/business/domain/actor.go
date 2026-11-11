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

// Authorize returns ErrForbidden unless the actor's role grants permission.
// An actor without ID is never authorized.
func (actor Actor) Authorize(permission Permission) error {
	if actor.ID == "" || !actor.Role.Can(permission) {
		return ErrForbidden
	}

	return nil
}
