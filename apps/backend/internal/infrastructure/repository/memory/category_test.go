package memory

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"market/backend/internal/business/gateway"
	"market/backend/internal/infrastructure/repository/repositorytest"
)

func TestCategoryRepositoryContract(t *testing.T) {
	repositorytest.CategoryRepository(t, func(*testing.T) gateway.CategoryRepository {
		return NewCategoryRepository(NewStore())
	})
}

func TestCategoryRepositoryHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewCategoryRepository(NewStore()).Get(ctx, "any")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Get() error = %v, want context.Canceled", err)
	}
}

func TestNewIDIsAUUIDv7(t *testing.T) {
	uuidV7 := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	first, second := newID(), newID()

	if !uuidV7.MatchString(first) {
		t.Fatalf("newID() = %q, want a UUIDv7", first)
	}
	if first == second {
		t.Fatal("newID() repeated an ID")
	}
}
