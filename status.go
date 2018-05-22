package multiserver

type status int

const (
	stopped status = iota
	started
	stopping
)
