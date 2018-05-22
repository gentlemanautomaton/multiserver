package multiserver

import "time"

// DefaultShutdownTimeout is the default amount of time that a graceful
// server shutdown will be allotted.
const DefaultShutdownTimeout = time.Second * 3

// Option is a group option.
type Option func(s *Group)

// ShutdownTimeout will use the given timeout when shutting a group of
// servers.
func ShutdownTimeout(timeout time.Duration) Option {
	return func(g *Group) {
		g.shutdownTimeout = timeout
	}
}
