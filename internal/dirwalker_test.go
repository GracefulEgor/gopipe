package internal

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestDirWalker(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "file.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	walker := NewDirWalker(filePath, 1, 0, logger)

	out := make(chan string, 1)
	walker.out[0] = out

	ctx := context.Background()
	go walker.Run(ctx)

	found := <- out
	if found != filePath {
		t.Errorf("expected %s, got %s", filePath, found)
	}

}