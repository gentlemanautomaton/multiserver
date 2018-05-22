package multiserver

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Group is a group of servers that will start and stop together.
type Group struct {
	servers         []Server
	shutdownTimeout time.Duration

	mutex  sync.Mutex
	status status
}

// New returns a new group of servers.
func New(servers ...Server) *Group {
	return &Group{servers: servers, shutdownTimeout: DefaultShutdownTimeout}
}

// Option applies the given options to the group.
func (g *Group) Option(opts ...Option) {
	for _, opt := range opts {
		opt(g)
	}
}

// ListenAndServe starts all servers within the group.
func (g *Group) ListenAndServe() error {
	if g.transition(stopped, started) {
		return g.execute(func(s Server) error {
			defer g.cancel()
			return s.ListenAndServe()
		})
	}

	return errors.New("server has already started")
}

// Shutdown initiates a graceful shutdown of all servers within the group.
func (g *Group) Shutdown(ctx context.Context) error {
	if g.transition(started, stopping) {
		defer g.transition(stopping, stopped)
		return g.execute(func(s Server) error {
			return s.Shutdown(ctx)
		})
	}

	return errors.New("server is already shutting down")
}

type task func(Server) error

func (g *Group) execute(t task) error {
	var (
		results = make([]error, len(g.servers))
		wg      sync.WaitGroup
	)

	// Prepare the waitgroup
	wg.Add(len(g.servers))

	// Start the servers
	for i, server := range g.servers {
		go func(wg *sync.WaitGroup, server Server, result *error) {
			defer wg.Done()
			*result = t(server)
		}(&wg, server, &results[i])
	}

	// Wait until all of the shutdown has completed or timed out for all
	// of the servers
	wg.Wait()

	// Return the first non-nil result
	for _, err := range results {
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Group) transition(from, to status) bool {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if g.status != from {
		return false
	}
	g.status = to
	return true
}

func (g *Group) cancel() {
	ctx, cancel := context.WithTimeout(context.Background(), g.shutdownTimeout)
	defer cancel()
	g.Shutdown(ctx)
}
