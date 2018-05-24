package multiserver

import "sync"

// collector collects the first err from a group.
type collector struct {
	once sync.Once
	err  error
}

func (c *collector) Apply(err error) {
	c.once.Do(func() {
		c.err = err
	})
}

func (c *collector) Result() error {
	return c.err
}
