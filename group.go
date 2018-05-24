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

// ListenAndServe starts all servers within the group by calling their
// ListenAndServe function. It returns only when each of these calls
// have returned.
//
// If the group has already been started ErrGroupStarted will be returned
// immediately.
func (g *Group) ListenAndServe() error {
	if g.transition(stopped, started) {
		return g.execute(func(s Server) error {
			return s.ListenAndServe()
		}, g.cancel)
	}

	return ErrGroupStarted
}

// ListenAndServeContext starts all servers within the group by calling their
// ListenAndServe function. It returns only when each of these calls
// have returned.
//
// If the group has already been started ErrGroupStarted will be returned
// immediately.
//
// Cancellation of the given context will result in a graceful shutdown of
// all servers. The amount of time allowed for each server to gracefully
// shutdown will be limited by the shutdown timeout of the group.
func (g *Group) ListenAndServeContext(ctx context.Context) error {
	if g.transition(stopped, started) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		var (
			started sync.WaitGroup
			result  = make(chan error)
		)

		started.Add(len(g.servers))

		go func() {
			defer close(result)
			result <- g.execute(func(s Server) error {
				go func() {
					time.Sleep(5 * time.Millisecond)
					started.Done()
				}()
				return s.ListenAndServe()
			}, g.cancel)
		}()

		started.Wait()

		go func() {
			<-ctx.Done()
			g.cancel()
		}()

		return <-result
	}

	return ErrGroupStarted
}

// Shutdown initiates a graceful shutdown of all servers within the group
// without interrupting any active connections. If the provided context
// expires before the shutdown is complete, Shutdown returns the context's
// error, otherwise it returns any error returned from closing the underlying
// net.Listener(s).
//
// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
// Make sure the program doesn't exit and waits instead for Shutdown to return.
func (g *Group) Shutdown(ctx context.Context) error {
	if g.transition(started, stopping) {
		defer g.transition(stopping, stopped)
		return g.execute(func(s Server) error {
			return s.Shutdown(ctx)
		}, nil)
	}

	return errors.New("server is already stopped or shutting down")
}

type task func(Server) error

func (g *Group) execute(t task, then func()) error {
	var (
		wg sync.WaitGroup
		c  collector
	)

	// Prepare the waitgroup
	wg.Add(len(g.servers))

	// Run the task for each server, and its followup if provided
	for _, server := range g.servers {
		go func(server Server) {
			defer wg.Done()
			if then != nil {
				defer then()
			}
			c.Apply(t(server))
		}(server)
	}

	// Wait until all of the tasks have completed
	wg.Wait()

	return c.Result()
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
