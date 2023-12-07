package connection

import "context"

type ConnectionID uint64

type connectionIDContextKey struct {
}

func AddConnectionIDToContext(
	ctx context.Context,
	connectionID ConnectionID,
) context.Context {
	key := connectionIDContextKey{}
	value := &connectionID

	return context.WithValue(ctx, key, value)
}

func GetConnectionIDFromContext(
	ctx context.Context,
) *ConnectionID {
	key := connectionIDContextKey{}

	if value := ctx.Value(key); value != nil {
		return value.(*ConnectionID)
	}

	return nil
}
