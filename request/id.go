package request

import (
	"context"
	"sync/atomic"
)

type RequestID uint64

func RequestIDFactory() func() RequestID {
	var previousRequestID atomic.Uint64

	return func() RequestID {
		return RequestID(previousRequestID.Add(1))
	}
}

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
) (requestID RequestID) {
	key := requestIDContextKey{}

	if id, ok := ctx.Value(key).(RequestID); ok {
		requestID = id
	}
	return
}
