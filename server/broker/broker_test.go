package broker

import (
	"sync"
	"testing"
)

func receive[T any](t *testing.T, sub *Subscription[T]) T {
	t.Helper()
	select {
	case message, ok := <-sub.C():
		if !ok {
			t.Fatal("subscription closed")
		}
		return message
	default:
		t.Fatal("no message")
	}
	panic("unreachable")
}

func TestPublishReachesEverySubscriber(t *testing.T) {
	b := New[int](4)
	first := b.Subscribe()
	second := b.Subscribe()

	b.Publish(1)
	b.Publish(2)

	for _, sub := range []*Subscription[int]{first, second} {
		if got := receive(t, sub); got != 1 {
			t.Errorf("first message = %d, want 1", got)
		}
		if got := receive(t, sub); got != 2 {
			t.Errorf("second message = %d, want 2", got)
		}
	}
}

func TestSubscriptionOnlyReceivesLaterMessages(t *testing.T) {
	b := New[int](4)
	b.Publish(1)
	sub := b.Subscribe()
	b.Publish(2)

	if got := receive(t, sub); got != 2 {
		t.Errorf("message = %d, want 2", got)
	}
}

func TestCloseUnsubscribes(t *testing.T) {
	b := New[int](4)
	sub := b.Subscribe()
	sub.Close()
	sub.Close()

	b.Publish(1)

	if _, ok := <-sub.C(); ok {
		t.Error("closed subscription received a message")
	}
	if sub.Dropped() {
		t.Error("closed subscription reports being dropped")
	}
	if got := b.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}
}

func TestPublishDropsFullSubscribers(t *testing.T) {
	b := New[int](1)
	slow := b.Subscribe()
	fast := b.Subscribe()

	b.Publish(1)
	if got := receive(t, fast); got != 1 {
		t.Fatalf("fast message = %d, want 1", got)
	}
	b.Publish(2)

	if got := receive(t, slow); got != 1 {
		t.Errorf("slow message = %d, want 1", got)
	}
	if _, ok := <-slow.C(); ok {
		t.Error("full subscription was not closed")
	}
	if !slow.Dropped() {
		t.Error("full subscription does not report being dropped")
	}
	if got := receive(t, fast); got != 2 {
		t.Errorf("fast message = %d, want 2", got)
	}
	slow.Close()
	if got := b.Len(); got != 1 {
		t.Errorf("Len() = %d, want 1", got)
	}
}

func TestConcurrentUse(t *testing.T) {
	b := New[int](1024)
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			sub := b.Subscribe()
			defer sub.Close()
			for range 100 {
				b.Publish(1)
			}
		})
	}
	wg.Wait()
	if got := b.Len(); got != 0 {
		t.Errorf("Len() = %d, want 0", got)
	}
}
