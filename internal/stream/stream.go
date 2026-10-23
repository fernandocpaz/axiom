// Package stream provides the cancellable event-stream primitives used by
// Axiom hosts such as TailA. Language values remain ordinary parser.Value
// values at this stage so the host runtime can be tested independently.
package stream

import (
	"context"
	"github.com/fernandocpaz/axiom/internal/parser"
	"sync"
)

type Event struct {
	Value parser.Value
	Err   error
}

type Source func(context.Context, chan<- Event)

type Stream struct{ source Source }

func New(source Source) Stream { return Stream{source: source} }

func Values(values ...parser.Value) Stream {
	return New(func(ctx context.Context, output chan<- Event) {
		for _, value := range values {
			select {
			case output <- Event{Value: value}:
			case <-ctx.Done():
				return
			}
		}
	})
}

func Errors(err error) Stream {
	return New(func(ctx context.Context, output chan<- Event) {
		select {
		case output <- Event{Err: err}:
		case <-ctx.Done():
		}
	})
}

func (s Stream) Map(transform func(parser.Value) (parser.Value, error)) Stream {
	return New(func(ctx context.Context, output chan<- Event) {
		input := s.events(ctx)
		defer input.stop()
		for event := range input.ch {
			if event.Err != nil {
				if !send(ctx, output, event) {
					return
				}
				continue
			}
			value, err := transform(event.Value)
			if !send(ctx, output, Event{Value: value, Err: err}) {
				return
			}
			if err != nil {
				return
			}
		}
	})
}

func (s Stream) Filter(keep func(parser.Value) (bool, error)) Stream {
	return New(func(ctx context.Context, output chan<- Event) {
		input := s.events(ctx)
		defer input.stop()
		for event := range input.ch {
			if event.Err != nil {
				if !send(ctx, output, event) {
					return
				}
				continue
			}
			matched, err := keep(event.Value)
			if err != nil {
				if !send(ctx, output, Event{Err: err}) {
					return
				}
				return
			}
			if matched && !send(ctx, output, event) {
				return
			}
		}
	})
}

func (s Stream) Take(limit int) Stream {
	return New(func(ctx context.Context, output chan<- Event) {
		if limit <= 0 {
			return
		}
		count := 0
		input := s.events(ctx)
		defer input.stop()
		for event := range input.ch {
			if !send(ctx, output, event) {
				return
			}
			if event.Err == nil {
				count++
			}
			if count >= limit {
				return
			}
		}
	})
}

func (s Stream) Buffer(capacity int) Stream {
	if capacity < 1 {
		capacity = 1
	}
	return New(func(ctx context.Context, output chan<- Event) {
		input := make(chan Event, capacity)
		upstream := s.events(ctx)
		defer upstream.stop()
		go func() {
			defer close(input)
			for event := range upstream.ch {
				select {
				case input <- event:
				case <-ctx.Done():
					return
				}
			}
		}()
		for event := range input {
			if !send(ctx, output, event) {
				return
			}
		}
	})
}

// BufferLatest keeps at most capacity events while allowing a live consumer
// to keep up with the newest data. When the queue is full, the oldest pending
// event is discarded. This is useful for TailA views where stale log lines
// are less valuable than current output.
func (s Stream) BufferLatest(capacity int) Stream {
	if capacity < 1 {
		capacity = 1
	}
	return New(func(ctx context.Context, output chan<- Event) {
		input := s.events(ctx)
		defer input.stop()

		queue := make([]Event, 0, capacity)
		for {
			var next chan<- Event
			var first Event
			if len(queue) > 0 {
				next = output
				first = queue[0]
			}

			select {
			case event, ok := <-input.ch:
				if !ok {
					for len(queue) > 0 {
						if !send(ctx, output, queue[0]) {
							return
						}
						queue = queue[1:]
					}
					return
				}
				if len(queue) == capacity {
					copy(queue, queue[1:])
					queue = queue[:capacity-1]
				}
				queue = append(queue, event)
			case next <- first:
				queue = queue[1:]
			case <-ctx.Done():
				return
			}
		}
	})
}

func Merge(sources ...Stream) Stream {
	return New(func(ctx context.Context, output chan<- Event) {
		var workers sync.WaitGroup
		workers.Add(len(sources))
		for _, source := range sources {
			go func(source Stream) {
				defer workers.Done()
				input := source.events(ctx)
				defer input.stop()
				for event := range input.ch {
					if !send(ctx, output, event) {
						return
					}
				}
			}(source)
		}
		workers.Wait()
	})
}

func (s Stream) Collect(ctx context.Context) ([]parser.Value, error) {
	var values []parser.Value
	input := s.events(ctx)
	defer input.stop()
	for event := range input.ch {
		if event.Err != nil {
			return nil, event.Err
		}
		values = append(values, event.Value)
	}
	return values, nil
}

type eventsResult struct {
	ch   <-chan Event
	stop func()
}

func (s Stream) events(ctx context.Context) eventsResult {
	output := make(chan Event)
	child, cancel := context.WithCancel(ctx)
	go func() {
		defer close(output)
		defer cancel()
		if s.source != nil {
			s.source(child, output)
		}
	}()
	return eventsResult{ch: output, stop: cancel}
}

func send(ctx context.Context, output chan<- Event, event Event) bool {
	select {
	case output <- event:
		return true
	case <-ctx.Done():
		return false
	}
}
