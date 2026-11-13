package customers

import (
	"context"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type ListCustomersInput struct {
	Actor  domain.Actor
	Search string
	Page   int
	Limit  int
}

type ListCustomers interface {
	Execute(ctx context.Context, input ListCustomersInput) (domain.Page[domain.Customer], error)
}

type listCustomers struct {
	repository gateway.CustomerRepository
}

func NewListCustomers(repository gateway.CustomerRepository) ListCustomers {
	return &listCustomers{repository: repository}
}

func (useCase *listCustomers) Execute(ctx context.Context, input ListCustomersInput) (domain.Page[domain.Customer], error) {
	if err := input.Actor.Authorize(domain.PermissionViewCustomers); err != nil {
		return domain.Page[domain.Customer]{}, err
	}

	page, err := domain.NewPageRequest(input.Page, input.Limit)
	if err != nil {
		return domain.Page[domain.Customer]{}, err
	}

	result, err := useCase.repository.List(ctx, gateway.CustomerFilter{
		Search: strings.TrimSpace(input.Search),
		Page:   page,
	})
	if err != nil {
		return domain.Page[domain.Customer]{}, domain.FromStorage(err)
	}

	return result, nil
}
