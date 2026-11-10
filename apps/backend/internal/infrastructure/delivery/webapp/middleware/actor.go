package middleware

import (
	"context"
	"net/http"

	"market/backend/internal/business/domain"
)

type actorKey struct{}

// DevActor attaches a fixed actor to every request. It stands in for real
// authentication until a provider is chosen; replacing it must not require
// changes in handlers or use cases, which only read ActorFromContext.
func DevActor(actor domain.Actor, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		next.ServeHTTP(w, request.WithContext(WithActor(request.Context(), actor)))
	})
}

func WithActor(ctx context.Context, actor domain.Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}

func ActorFromContext(ctx context.Context) (domain.Actor, bool) {
	actor, ok := ctx.Value(actorKey{}).(domain.Actor)

	return actor, ok
}
