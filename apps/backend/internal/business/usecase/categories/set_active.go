package categories

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// setActive is shared by DeactivateCategory and ReactivateCategory. Asking
// for the state a category already has returns it unchanged, so a retried
// request succeeds.
func setActive(ctx context.Context, repository gateway.CategoryRepository, clock gateway.Clock, actor domain.Actor, id string, active bool) (domain.Category, error) {
	if err := actor.Authorize(manageRoles...); err != nil {
		return domain.Category{}, err
	}

	current, err := repository.Get(ctx, id)
	if err != nil {
		return domain.Category{}, domain.FromStorage(err)
	}
	if current.Active == active {
		return current, nil
	}

	updated, err := repository.Update(ctx, current.WithActive(active, clock.Now()))
	if err != nil {
		return domain.Category{}, domain.FromStorage(err)
	}

	return updated, nil
}
