package main

import (
	"sync"
	"time"
)

type EventStats interface {
	UpdateStat(stat *Stat, e *Event)
	Subscribe() chan *Event
	Unsubscribe(chan *Event)
}

type SimpleEventStats struct {
	mu          sync.Mutex
	subscribers map[chan *Event]struct{}
	stats       Stat
}

func (ss *SimpleEventStats) InitStat() *Stat {
	return &Stat{
		ByMethod:   make(map[string]uint64),
		ByConsumer: make(map[string]uint64),
	}
}

func (ss *SimpleEventStats) UpdateStat(e *Event) {
	ss.stats.Timestamp = time.Now().Unix()
	ss.stats.ByConsumer[e.Consumer]++
	ss.stats.ByMethod[e.Method]++
}

func (ss *SimpleEventStats) Subscribe() chan *Event {
	ch := make(chan *Event)
	ss.mu.Lock()

	defer ss.mu.Unlock()

	ss.subscribers[ch] = struct{}{}

	return ch
}

func (ss *SimpleEventStats) Unsubscribe(ch chan *Event) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.subscribers, ch)
}
