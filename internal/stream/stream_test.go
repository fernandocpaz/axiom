package stream

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fernandocpaz/axiom/internal/parser"
)

func TestPipelineAndTake(t *testing.T) {
	values, err := Values(
		parser.Value{Kind: "int", Int: 1},
		parser.Value{Kind: "int", Int: 2},
		parser.Value{Kind: "int", Int: 3},
	).Filter(func(value parser.Value) (bool, error) { return value.Int > 1, nil }).Take(1).Collect(context.Background())
	if err != nil || len(values) != 1 || values[0].Int != 2 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}

func TestCancellationStopsSource(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	source := New(func(ctx context.Context, output chan<- Event) {
		close(started)
		<-ctx.Done()
	})
	result := make(chan error, 1)
	go func() { _, err := source.Collect(ctx); result <- err }()
	<-started
	cancel()
	if err := <-result; err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTakeCancelsUpstream(t *testing.T) {
	started := make(chan struct{})
	stopped := make(chan struct{})
	source := New(func(ctx context.Context, output chan<- Event) {
		close(started)
		defer close(stopped)
		for {
			select {
			case output <- Event{Value: parser.Value{Kind: "int", Int: 1}}:
			case <-ctx.Done():
				return
			}
		}
	})

	values, err := source.Take(1).Collect(context.Background())
	if err != nil || len(values) != 1 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("upstream source was not cancelled")
	}
}

func TestErrorsPropagate(t *testing.T) {
	expected := errors.New("broken source")
	_, err := Errors(expected).Collect(context.Background())
	if !errors.Is(err, expected) {
		t.Fatalf("got %v, want %v", err, expected)
	}
}

func TestMergeCombinesSources(t *testing.T) {
	values, err := Merge(
		Values(parser.Value{Kind: "string", Str: "a"}),
		Values(parser.Value{Kind: "string", Str: "b"}),
	).Buffer(2).Collect(context.Background())
	if err != nil || len(values) != 2 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}

func TestBufferLatestDropsOldestPendingEvents(t *testing.T) {
	gate := make(chan struct{})
	ready := make(chan struct{})
	source := New(func(ctx context.Context, output chan<- Event) {
		for i := 1; i <= 4; i++ {
			select {
			case output <- Event{Value: parser.Value{Kind: "int", Int: i}}:
			case <-ctx.Done():
				return
			}
		}
		close(ready)
		<-gate
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	input := source.BufferLatest(2).events(ctx)
	defer input.stop()
	select {
	case <-ready:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for source to fill buffer")
	}

	select {
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for buffered events")
	case event := <-input.ch:
		if event.Value.Int != 3 {
			t.Fatalf("first event=%d, want newest queued event 3", event.Value.Int)
		}
	}
	select {
	case event := <-input.ch:
		if event.Value.Int != 4 {
			t.Fatalf("second event=%d, want newest queued event 4", event.Value.Int)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for second buffered event")
	}
	close(gate)
}
