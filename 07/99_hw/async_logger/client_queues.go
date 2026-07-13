package main

import (
	sync "sync"

	ulid "github.com/oklog/ulid/v2"
)

type ClientQueues struct {
	channels map[ulid.ULID]chan *Event
	mu       *sync.Mutex
}

func NewClientQueues() ClientQueues {
	return ClientQueues{
		channels: make(map[ulid.ULID]chan *Event),
		mu:       &sync.Mutex{},
	}
}

func (c *ClientQueues) Connect() (ulid.ULID, chan *Event) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := ulid.Make()
	ch := make(chan *Event, 1)
	c.channels[id] = ch

	return id, ch
}

func (c *ClientQueues) Disconnect(id ulid.ULID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	close(c.channels[id])
	delete(c.channels, id)
}
