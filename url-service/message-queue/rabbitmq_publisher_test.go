package messagequeue

import (
	"sync"
	"testing"
)

func TestRabbitmqPublisherPublishNilConnection(t *testing.T) {
	publisher := NewRabbitmqPublisher(nil)

	err := publisher.Publish(RoutingKeyLinkClicked, []byte(`{"event_id":"x"}`))
	if err != ErrQueueUnavailable {
		t.Fatalf("expected ErrQueueUnavailable, got %v", err)
	}
}

func TestRabbitmqPublisherConcurrentNilConnection(t *testing.T) {
	publisher := NewRabbitmqPublisher(nil)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := publisher.Publish(RoutingKeyLinkClicked, []byte(`{}`)); err != ErrQueueUnavailable {
				t.Errorf("expected ErrQueueUnavailable, got %v", err)
			}
		}()
	}

	wg.Wait()
}