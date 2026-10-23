package stream

import (
	"context"
	"errors"
	"testing"

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
