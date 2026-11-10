// Package broker fans messages out to subscribers in memory.
package broker

import "sync"

// Broker delivers each published message to every current subscriber.
//
// Publish never blocks: each subscriber has a buffer, and a subscriber whose
// buffer is full is dropped, so one slow subscriber cannot hold up the
// publisher or the others. The zero value is not usable; call New.
type Broker[T any] struct {
	mu          sync.Mutex
	buffer      int
	subscribers map[*Subscription[T]]struct{}
}

// Subscription receives the messages published after it was created.
type Subscription[T any] struct {
	broker  *Broker[T]
	ch      chan T
	dropped bool
}

// New returns a Broker whose subscribers buffer up to buffer messages.
func New[T any](buffer int) *Broker[T] {
	return &Broker[T]{
		buffer:      max(buffer, 1),
		subscribers: map[*Subscription[T]]struct{}{},
	}
}

// Subscribe returns a subscription to the messages published from now on.
// Close it once it is no longer read.
func (b *Broker[T]) Subscribe() *Subscription[T] {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := &Subscription[T]{broker: b, ch: make(chan T, b.buffer)}
	b.subscribers[sub] = struct{}{}
	return sub
}

// Publish delivers message to every subscriber, and drops each subscriber
// whose buffer is full.
func (b *Broker[T]) Publish(message T) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for sub := range b.subscribers {
		select {
		case sub.ch <- message:
		default:
			sub.dropped = true
			b.remove(sub)
		}
	}
}

// Len returns the number of subscribers.
func (b *Broker[T]) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return len(b.subscribers)
}

// remove closes sub's channel and forgets it. b.mu must be held.
func (b *Broker[T]) remove(sub *Subscription[T]) {
	if _, ok := b.subscribers[sub]; !ok {
		return
	}
	delete(b.subscribers, sub)
	close(sub.ch)
}

// C returns the channel that receives the messages. It is closed when the
// subscription is closed or dropped.
func (s *Subscription[T]) C() <-chan T {
	return s.ch
}

// Dropped reports whether the broker dropped the subscription because its
// buffer was full. It is only meaningful once C is closed.
func (s *Subscription[T]) Dropped() bool {
	s.broker.mu.Lock()
	defer s.broker.mu.Unlock()

	return s.dropped
}

// Close unsubscribes. It is safe to call more than once, and after the
// subscription was dropped.
func (s *Subscription[T]) Close() {
	s.broker.mu.Lock()
	defer s.broker.mu.Unlock()

	s.broker.remove(s)
}
