package http

import (
	"context"

	"github.com/avito-internships/test-backend-1-EdOoO21/internal/application/shared"
)

type actorContextKey struct{}
type servicesContextKey struct{}

func withActor(ctx context.Context, actor shared.Actor) context.Context {
	return context.WithValue(ctx, actorContextKey{}, actor)
}

func actorFromContext(ctx context.Context) (shared.Actor, bool) {
	actor, ok := ctx.Value(actorContextKey{}).(shared.Actor)
	return actor, ok
}

func withServices(ctx context.Context, services Services) context.Context {
	return context.WithValue(ctx, servicesContextKey{}, services)
}

func servicesFromContext(ctx context.Context) (Services, bool) {
	services, ok := ctx.Value(servicesContextKey{}).(Services)
	return services, ok
}
