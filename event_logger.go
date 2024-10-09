package main

import (
	"sync"
	"time"
)

type EventLogger interface {
	LogEvent(consumer, method, host string)
	Subscribe() chan *Event
	Unsubscribe(chan *Event)
}

type SimpleEventLogger struct {
	mu          sync.Mutex
	subscribers map[chan *Event]struct{}
}

func (el *SimpleEventLogger) LogEvent(consumer, method, host string) {
	e := &Event{
		Timestamp: time.Now().Unix(),
		Consumer:  consumer,
		Method:    method,
		Host:      host,
	}
	el.mu.Lock()
	defer el.mu.Unlock()
	for sub := range el.subscribers {
		sub <- e
	}
}

func (el *SimpleEventLogger) Subscribe() chan *Event {
	ch := make(chan *Event)
	el.mu.Lock()

	defer el.mu.Unlock()

	el.subscribers[ch] = struct{}{}

	return ch
}

func (el *SimpleEventLogger) Unsubscribe(ch chan *Event) {
	el.mu.Lock()
	defer el.mu.Unlock()
	delete(el.subscribers, ch)
}
