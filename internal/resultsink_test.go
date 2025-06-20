package internal

import (
	"context"
	"log/slog"
	"os"
	"testing"
)

func TestResultSink_ReceivesResults(t *testing.T) {
	in := make(chan FileHash, 1)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	sink := NewResultSink(1, logger)
	sink.in[0] = in

	ctx := context.Background()
	go func() {
		in <- FileHash{Path: "foo", Hash: "bar", Err: nil}
		close(in)
	}()

	err := sink.Run(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
