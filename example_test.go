package multiserver_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gentlemanautomaton/multiserver"
)

func Example() {
	// Prepare server A
	s1 := &http.Server{
		Addr:    ":8080",
		Handler: respondWith("one"),
	}

	// Prepare server B
	s2 := &http.Server{
		Addr:    ":8081",
		Handler: respondWith("two"),
	}

	// Prepare the group
	s := multiserver.New(s1, s2)

	// If a server dies unexpectedly, give the others one second to
	// gracefully shut down.
	s.Option(multiserver.ShutdownTimeout(time.Second))

	// Start the servers
	go s.ListenAndServe()

	// Wait for the servers to start
	time.Sleep(10 * time.Millisecond)

	// Query each server
	fmt.Println(query("http://localhost:8080/"))
	fmt.Println(query("http://localhost:8081/"))

	// Tell the servers to perform a graceful shutdown without any timeout
	s.Shutdown(context.Background())

	// Output:
	// one
	// two
}

func respondWith(value string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprint(w, value)
	}
}

func query(url string) string {
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	return string(body)
}
