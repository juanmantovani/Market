package categories

import (
	"context"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type ListCategoriesInput struct {
	Actor  domain.Actor
	Search string
	Status string
	Page   int
	Limit  int
}

type ListCategories interface {
	Execute(ctx context.Context, input ListCategoriesInput) (domain.Page[domain.Category], error)
}

type listCategories struct {
	repository gateway.CategoryRepository
}

func NewListCategories(repository gateway.CategoryRepository) ListCategories {
	return &listCategories{repository: repository}
}

func (useCase *listCategories) Execute(ctx context.Context, input ListCategoriesInput) (domain.Page[domain.Category], error) {
	if err := input.Actor.Authorize(readRoles...); err != nil {
		return domain.Page[domain.Category]{}, err
	}

	status, err := domain.ParseStatusFilter(input.Status)
	if err != nil {
		return domain.Page[domain.Category]{}, err
	}
	page, err := domain.NewPageRequest(input.Page, input.Limit)
	if err != nil {
		return domain.Page[domain.Category]{}, err
	}

	result, err := useCase.repository.List(ctx, gateway.CategoryFilter{
		Search: strings.TrimSpace(input.Search),
		Status: status,
		Page:   page,
	})
	if err != nil {
		return domain.Page[domain.Category]{}, domain.FromStorage(err)
	}

	return result, nil
}
