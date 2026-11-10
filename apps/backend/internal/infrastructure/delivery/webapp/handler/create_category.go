package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/categories"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/categories"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type CreateCategoryHandler struct {
	createCategory categories.CreateCategory
}

func NewCreateCategoryHandler(createCategory categories.CreateCategory) *CreateCategoryHandler {
	return &CreateCategoryHandler{createCategory: createCategory}
}

// Handle serves POST /api/v1/categories.
func (handler *CreateCategoryHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.CreateCategoryRequest](request)
	if err != nil {
		return err
	}

	category, err := handler.createCategory.Execute(request.Context(), categories.CreateCategoryInput{
		Actor:       actorFrom(request),
		Name:        body.Name,
		Description: body.Description,
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusCreated, dto.NewCategoryResponse(category))
	return nil
}
