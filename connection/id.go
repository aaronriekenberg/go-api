package connection

import (
	"context"
	"strconv"
)

type ConnectionID uint64

func (connectionID *ConnectionID) String() string {
	if connectionID == nil {
		return "(nil)"
	}

	return strconv.FormatUint(uint64(*connectionID), 10)
}

type connectionIDContextKey struct{}

func AddConnectionIDToContext(
	ctx context.Context,
	connectionID ConnectionID,
) context.Context {
	key := connectionIDContextKey{}
	value := connectionID

	return context.WithValue(ctx, key, value)
}

func GetConnectionIDFromContext(
	ctx context.Context,
) *ConnectionID {
	key := connectionIDContextKey{}

	if value, ok := ctx.Value(key).(ConnectionID); ok {
		return &value
	}

	return nil
}
