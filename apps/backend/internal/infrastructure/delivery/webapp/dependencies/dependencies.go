package dependencies

import (
	"fmt"

	"market/backend/internal/business/gateway"
	"market/backend/internal/business/usecase/categories"
	"market/backend/internal/business/usecase/items"
	"market/backend/internal/infrastructure/delivery/webapp/handler"
	"market/backend/internal/infrastructure/delivery/webapp/route"
	"market/backend/internal/infrastructure/environments"
	"market/backend/internal/infrastructure/repository/memory"
)

// Container is the object graph: repositories → use cases → handlers. It
// only builds objects; reading configuration, registering routes and running
// the server belong to internal/app.
type Container struct {
	CreateCategory categories.CreateCategory
	ListCategories categories.ListCategories
	GetCategory    categories.GetCategory
	UpdateCategory categories.UpdateCategory
	DeleteCategory categories.DeleteCategory

	CreateItem items.CreateItem
	ListItems  items.ListItems
	GetItem    items.GetItem
	UpdateItem items.UpdateItem
	DeleteItem items.DeleteItem
}

type repositories struct {
	categories gateway.CategoryRepository
	items      gateway.ItemRepository
}

func New(cfg environments.Server) (*Container, error) {
	repos, err := newRepositories(cfg.StorageDriver)
	if err != nil {
		return nil, err
	}

	return &Container{
		CreateCategory: categories.NewCreateCategory(repos.categories),
		ListCategories: categories.NewListCategories(repos.categories),
		GetCategory:    categories.NewGetCategory(repos.categories),
		UpdateCategory: categories.NewUpdateCategory(repos.categories),
		DeleteCategory: categories.NewDeleteCategory(repos.categories),

		CreateItem: items.NewCreateItem(repos.items),
		ListItems:  items.NewListItems(repos.items),
		GetItem:    items.NewGetItem(repos.items),
		UpdateItem: items.NewUpdateItem(repos.items),
		DeleteItem: items.NewDeleteItem(repos.items),
	}, nil
}

func newRepositories(driver environments.StorageDriver) (repositories, error) {
	switch driver {
	case environments.StorageMemory:
		store := memory.NewStore()
		return repositories{
			categories: memory.NewCategoryRepository(store),
			items:      memory.NewItemRepository(store),
		}, nil
	default:
		return repositories{}, fmt.Errorf("unsupported storage driver %q", driver)
	}
}

// Handlers wraps every use case in its route's handler.
func (container *Container) Handlers() route.Handlers {
	return route.Handlers{
		ListCategories: handler.NewListCategoriesHandler(container.ListCategories),
		GetCategory:    handler.NewGetCategoryHandler(container.GetCategory),
		CreateCategory: handler.NewCreateCategoryHandler(container.CreateCategory),
		UpdateCategory: handler.NewUpdateCategoryHandler(container.UpdateCategory),
		DeleteCategory: handler.NewDeleteCategoryHandler(container.DeleteCategory),

		ListItems:  handler.NewListItemsHandler(container.ListItems),
		GetItem:    handler.NewGetItemHandler(container.GetItem),
		CreateItem: handler.NewCreateItemHandler(container.CreateItem),
		UpdateItem: handler.NewUpdateItemHandler(container.UpdateItem),
		DeleteItem: handler.NewDeleteItemHandler(container.DeleteItem),
	}
}
