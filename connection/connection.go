package connection

import (
	"sync/atomic"
	"time"
)

type Connection interface {
	ID() ConnectionID
	CreationTime() time.Time
	Age(now time.Time) time.Duration
	Requests() uint64
}

type connection struct {
	id           ConnectionID
	creationTime time.Time
	requests     atomic.Uint64
	closeTime    time.Time
}

func newConnection(
	id ConnectionID,
) *connection {
	return &connection{
		id:           id,
		creationTime: time.Now(),
	}
}

func (c *connection) ID() ConnectionID {
	return c.id
}

func (c *connection) CreationTime() time.Time {
	return c.creationTime
}

func (c *connection) Age(now time.Time) time.Duration {
	return now.Sub(c.creationTime)
}

func (c *connection) Requests() uint64 {
	return c.requests.Load()
}

func (c *connection) incrementRequests() {
	c.requests.Add(1)
}

func (c *connection) markClosed() {
	c.closeTime = time.Now()
}

func (c *connection) openDuration() time.Duration {
	return c.Age(c.closeTime)
}
