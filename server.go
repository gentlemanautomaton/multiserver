package multiserver

import "context"

// Server is an interface of something that can be started and shutdown as
// part of a group.
type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}
