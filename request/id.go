package request

import (
	"context"
	"sync/atomic"
)

type RequestID int

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

var previousRequestID atomic.Int64

func NextRequestID() RequestID {
	id := previousRequestID.Add(1)
	return RequestID(id)
}
