package request

import (
	"context"
	"strconv"
	"sync/atomic"
)

type RequestID uint64

func (requestID *RequestID) String() string {
	if requestID == nil {
		return "(nil)"
	}

	return strconv.FormatUint(uint64(*requestID), 10)
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
) *RequestID {
	key := requestIDContextKey{}

	if value, ok := ctx.Value(key).(RequestID); ok {
		return &value
	}

	return nil
}

var nextRequestID atomic.Uint64

func NextRequestID() RequestID {
	id := nextRequestID.Add(1)
	return RequestID(id)
}
