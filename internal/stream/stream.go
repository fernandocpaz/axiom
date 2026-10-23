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
		for event := range s.events(ctx) {
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
		for event := range s.events(ctx) {
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
		for event := range s.events(ctx) {
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
		go func() {
			defer close(input)
			for event := range s.events(ctx) {
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

func Merge(sources ...Stream) Stream {
	return New(func(ctx context.Context, output chan<- Event) {
		var workers sync.WaitGroup
		workers.Add(len(sources))
		for _, source := range sources {
			go func(source Stream) {
				defer workers.Done()
				for event := range source.events(ctx) {
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
	for event := range s.events(ctx) {
		if event.Err != nil {
			return nil, event.Err
		}
		values = append(values, event.Value)
	}
	return values, nil
}

func (s Stream) events(ctx context.Context) <-chan Event {
	output := make(chan Event)
	go func() {
		defer close(output)
		if s.source != nil {
			s.source(ctx, output)
		}
	}()
	return output
}

func send(ctx context.Context, output chan<- Event, event Event) bool {
	select {
	case output <- event:
		return true
	case <-ctx.Done():
		return false
	}
}
