package internal

import (
	"context"
	"log/slog"
	"os"
	"testing"
)


func TestMD5Worker(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "md5test")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	tmpFile.WriteString("hello")
	tmpFile.Close()

	in := make(chan string, 1)
	out := make(chan FileHash, 1)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	worker := NewMD5Worker(1, 1, 1, logger)

	worker.in[0] = in
	worker.out[0] = out

	ctx := context.Background()
	in <- tmpFile.Name()
	close(in)

	go worker.Run(ctx)

	res := <- out
	expected := "5d41402abc4b2a76b9719d911017c592"
	if res.Hash != expected {
		t.Errorf("expected hash %s, got %s", expected, res.Hash)
	}
	if res.Err != nil {
		t.Errorf("unexpected error: %v", res.Err)
	}
}
