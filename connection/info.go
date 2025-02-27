package connection

import (
	"context"
	"sync/atomic"
	"time"
)

type ConnectionID uint64

type ConnectionInfo interface {
	ID() ConnectionID
	Network() string
	CreationTime() time.Time
	Age(now time.Time) time.Duration
	Requests() uint64
	IncrementRequests()
	markClosed()
	openDuration() time.Duration
}

type connectionInfo struct {
	id           ConnectionID
	network      string
	creationTime time.Time
	requests     atomic.Uint64
	closeTime    time.Time
}

func newConnection(
	id ConnectionID,
	network string,
) ConnectionInfo {
	return &connectionInfo{
		id:           id,
		network:      network,
		creationTime: time.Now(),
	}
}

func (ci *connectionInfo) ID() ConnectionID {
	return ci.id
}

func (ci *connectionInfo) Network() string {
	return ci.network
}

func (ci *connectionInfo) CreationTime() time.Time {
	return ci.creationTime
}

func (ci *connectionInfo) Age(now time.Time) time.Duration {
	return now.Sub(ci.creationTime)
}

func (ci *connectionInfo) Requests() uint64 {
	return ci.requests.Load()
}

func (ci *connectionInfo) IncrementRequests() {
	ci.requests.Add(1)
}

func (ci *connectionInfo) markClosed() {
	ci.closeTime = time.Now()
}

func (ci *connectionInfo) openDuration() time.Duration {
	return ci.Age(ci.closeTime)
}

type connectionInfoContextKey struct{}

func AddConnectionInfoToContext(
	ctx context.Context,
	connectionInfo ConnectionInfo,
) context.Context {
	key := connectionInfoContextKey{}
	value := connectionInfo

	return context.WithValue(ctx, key, value)
}

func ConnectionInfoFromContext(
	ctx context.Context,
) (connectionInfo ConnectionInfo, ok bool) {
	key := connectionInfoContextKey{}

	connectionInfo, ok = ctx.Value(key).(ConnectionInfo)
	return
}

func ConnectionIDFromContext(
	ctx context.Context,
) (connectionID ConnectionID) {
	if connectionInfo, ok := ConnectionInfoFromContext(ctx); ok {
		connectionID = connectionInfo.ID()
	}
	return
}
