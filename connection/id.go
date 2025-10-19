package connection

import "sync/atomic"

type ConnectionID uint64

func connectionIDFactory() func() ConnectionID {
	var previousConnectionID atomic.Uint64

	return func() ConnectionID {
		return ConnectionID(previousConnectionID.Add(1))
	}
}
