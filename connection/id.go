package connection

import (
	"context"
)

type ConnectionID uint64

type connectionIDContextKey struct{}

func AddConnectionIDToContext(
	ctx context.Context,
	connectionID ConnectionID,
) context.Context {
	key := connectionIDContextKey{}
	value := connectionID

	return context.WithValue(ctx, key, value)
}

func ConnectionIDFromContext(
	ctx context.Context,
) (connectionID ConnectionID, ok bool) {
	key := connectionIDContextKey{}

	connectionID, ok = ctx.Value(key).(ConnectionID)
	return
}
