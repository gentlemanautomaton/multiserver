package multiserver_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gentlemanautomaton/multiserver"
)

var testError = errors.New("test err")

type testServer struct {
	wait  bool
	delay time.Duration
	err   error

	startup  sync.Once
	shutdown sync.Once
	done     chan struct{}
}

func (srv *testServer) ListenAndServe() error {
	if srv.delay != 0 {
		time.Sleep(srv.delay)
	}
	if srv.wait {
		srv.startup.Do(srv.init)
		<-srv.done
		return http.ErrServerClosed
	}
	return srv.err
}

func (srv *testServer) Shutdown(ctx context.Context) error {
	srv.startup.Do(srv.init)
	srv.shutdown.Do(func() {
		close(srv.done)
	})
	return nil
}

func (srv *testServer) init() {
	srv.done = make(chan struct{})
}

type testServerSet []testServer

func (s testServerSet) RunListenAndServe(t *testing.T, expected error) {
	t.Parallel()
	servers := make([]multiserver.Server, len(s))
	for i := range s {
		servers[i] = &s[i]
	}
	g := multiserver.New(servers...)
	actual := g.ListenAndServe()
	if actual != expected {
		t.Errorf("actual \"%v\", expected \"%v\"", actual, expected)
	}
}

func TestListenAndServe(t *testing.T) {
	expected := testError
	for round := 0; round < 5; round++ {
		delay := 5 * time.Millisecond * time.Duration(round)
		t.Run(fmt.Sprintf("delay %s", delay), func(t *testing.T) {
			tss := testServerSet{
				testServer{delay: delay, err: expected},
				testServer{delay: 10 * time.Millisecond, wait: true},
			}
			tss.RunListenAndServe(t, expected)
		})
	}
}

func TestListenAndServeStarted(t *testing.T) {
	g := multiserver.New(&testServer{delay: 30 * time.Millisecond})
	go g.ListenAndServe()
	time.Sleep(10 * time.Millisecond)
	actual := g.ListenAndServe()
	expected := multiserver.ErrGroupStarted
	if actual != expected {
		t.Errorf("actual \"%v\", expected \"%v\"", actual, expected)
	}
}

func TestListenAndServeContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	g := multiserver.New(&testServer{wait: true})
	actual := g.ListenAndServeContext(ctx)
	expected := http.ErrServerClosed
	if actual != expected {
		t.Errorf("actual \"%v\", expected \"%v\"", actual, expected)
	}
}

func TestListenAndServeContextStarted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	g := multiserver.New(&testServer{delay: 30 * time.Millisecond})
	go g.ListenAndServeContext(ctx)
	time.Sleep(10 * time.Millisecond)
	actual := g.ListenAndServeContext(ctx)
	expected := multiserver.ErrGroupStarted
	if actual != expected {
		t.Errorf("actual \"%v\", expected \"%v\"", actual, expected)
	}
}
