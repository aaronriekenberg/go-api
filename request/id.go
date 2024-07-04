package request

import (
	"context"
	"sync/atomic"
)

type RequestID uint64

type requestIDContextKey struct{}

func AddRequestIDToContext(
	ctx context.Context,
	requestID RequestID,
) context.Context {
	key := requestIDContextKey{}
	value := requestID

	return context.WithValue(ctx, key, value)
}

func RequestIDFromContext(
	ctx context.Context,
) (requestID RequestID, ok bool) {
	key := requestIDContextKey{}

	requestID, ok = ctx.Value(key).(RequestID)
	return
}

var previousRequestID atomic.Uint64

func NextRequestID() RequestID {
	id := previousRequestID.Add(1)
	return RequestID(id)
}
