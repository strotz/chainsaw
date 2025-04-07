package link

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/strotz/chainsaw/link/def"
)

// structure to demonstrate the async callbacks
type sampleBus struct {
	callback func()
}

func (e *sampleBus) subscribe(callback func()) {
	e.callback = callback
}
func (e *sampleBus) notify() {
	if e.callback != nil {
		go e.callback()
	}
}

type source interface {
	GetCallId() def.CallId
}

// Test the subscribe and notify methods of the sampleBus are working
func TestSubscribe(t *testing.T) {
	bus := &sampleBus{}
	var wg sync.WaitGroup
	wg.Add(1)
	bus.subscribe(func() {
		wg.Done()
	})
	go bus.notify()
	wg.Wait()
}

func TestCreateJob(t *testing.T) {
	bus := &sampleBus{}
	// The typical usage of the jobi is to send a message to the bus and wait for a response.
	// Question: how to link the response? it should be some sort of transaction id.
	// question: what is the timeout methodology? context? or bus? or this function has a way to define a timeout?
	x := func(ctx context.Context) (string, error) {
		result := make(chan string)
		bus.subscribe(func() {
			// It feels like it has to be a list of subscribers.
			result <- "result"
		})
		bus.notify()
		// It waits for the response or the context to be done
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case r, ok := <-result:
			if !ok {
				return "", errors.New("channel closed")
			}
			return r, nil
		}
	}

	ctx := context.TODO()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	_, err := x(ctx)
	require.NoError(t, err)

	//job := CreateJob(ctx)
	//r, err :=
}
