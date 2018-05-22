package multiserver

import "time"

// DefaultShutdownTimeout is the default amount of time that a graceful
// server shutdown will be allotted in the case of an unexpected failure.
const DefaultShutdownTimeout = time.Second * 3

// Option is a group option.
type Option func(s *Group)

// ShutdownTimeout defines the amount of time alloted for a group of servers
// to gracefully shutdown when one of their peers stops unexpectedly.
//
// If group.Shutdown() is called directly this value will be ignored.
func ShutdownTimeout(timeout time.Duration) Option {
	return func(g *Group) {
		g.shutdownTimeout = timeout
	}
}
