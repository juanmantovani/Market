package memory

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

var _ gateway.CategoryRepository = (*CategoryRepository)(nil)

type CategoryRepository struct {
	store *Store
}

func NewCategoryRepository(store *Store) *CategoryRepository {
	return &CategoryRepository{store: store}
}

func (repository *CategoryRepository) Create(ctx context.Context, category domain.Category) (domain.Category, error) {
	var created domain.Category
	err := repository.store.run(ctx, func(current *tx) error {
		category.ID = newID()
		if err := repository.checkNameIsFree(category); err != nil {
			return err
		}
		put(current, repository.store.categories, category.ID, cloneCategory(category))
		created = cloneCategory(category)
		return nil
	})

	return created, err
}

func (repository *CategoryRepository) Get(ctx context.Context, id string) (domain.Category, error) {
	var found domain.Category
	err := repository.store.run(ctx, func(*tx) error {
		category, ok := repository.store.categories[id]
		if !ok {
			return domain.ErrCategoryNotFound
		}
		found = cloneCategory(category)
		return nil
	})

	return found, err
}

func (repository *CategoryRepository) Update(ctx context.Context, category domain.Category) (domain.Category, error) {
	var updated domain.Category
	err := repository.store.run(ctx, func(current *tx) error {
		if _, ok := repository.store.categories[category.ID]; !ok {
			return domain.ErrCategoryNotFound
		}
		if err := repository.checkNameIsFree(category); err != nil {
			return err
		}
		put(current, repository.store.categories, category.ID, cloneCategory(category))
		updated = cloneCategory(category)
		return nil
	})

	return updated, err
}

func (repository *CategoryRepository) List(ctx context.Context, filter gateway.CategoryFilter) (domain.Page[domain.Category], error) {
	page := domain.Page[domain.Category]{Items: []domain.Category{}, Page: filter.Page.Page, Limit: filter.Page.Limit}
	err := repository.store.run(ctx, func(*tx) error {
		search := domain.NormalizeText(filter.Search)
		var matches []domain.Category
		for _, category := range repository.store.categories {
			if filter.Status.Matches(category.Active) && categoryMatches(category, search) {
				matches = append(matches, category)
			}
		}
		slices.SortFunc(matches, func(a, b domain.Category) int {
			return cmp.Or(cmp.Compare(a.NameKey(), b.NameKey()), cmp.Compare(a.ID, b.ID))
		})

		page.Total = len(matches)
		start := min(filter.Page.Offset(), len(matches))
		end := min(start+filter.Page.Limit, len(matches))
		for _, category := range matches[start:end] {
			page.Items = append(page.Items, cloneCategory(category))
		}
		return nil
	})

	return page, err
}

// checkNameIsFree enforces what a partial unique index on the normalized
// name of active categories enforces in PostgreSQL.
func (repository *CategoryRepository) checkNameIsFree(category domain.Category) error {
	if !category.Active {
		return nil
	}
	key := category.NameKey()
	for id, other := range repository.store.categories {
		if id != category.ID && other.Active && other.NameKey() == key {
			return domain.ErrCategoryNameInUse
		}
	}

	return nil
}

func categoryMatches(category domain.Category, search string) bool {
	if search == "" {
		return true
	}
	if strings.Contains(category.NameKey(), search) {
		return true
	}

	return category.Description != nil && strings.Contains(domain.NormalizeText(*category.Description), search)
}

func cloneCategory(category domain.Category) domain.Category {
	if category.Description != nil {
		description := *category.Description
		category.Description = &description
	}

	return category
}
