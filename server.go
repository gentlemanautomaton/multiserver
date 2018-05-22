package multiserver

import "context"

// Server is server that can be started and shutdown as part of a group.
type Server interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}
