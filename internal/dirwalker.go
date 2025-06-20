package internal

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
)

type DirWalker struct {
	out []chan<- string
	root string
	logger *slog.Logger
}

func NewDirWalker(root string, numOut, numIn int, logger *slog.Logger) *DirWalker {
	out := make([]chan<- string, numOut)
	return &DirWalker{
		out: out,
		root: root,
		logger: logger,
	}
}

func (w *DirWalker) In() []<-chan string {return nil}

func (w *DirWalker) Out() []chan<- string {return w.out}


func (w *DirWalker) Run(ctx context.Context) error {
	w.logger.Info("dirwlkr started", "root", w.root)
	defer func() {
		for _, ch := range w.out {
			close(ch)
		}
		w.logger.Info("dirwlkr finished", "root", w.root)
	}() 

	walkFn := func(path string, d os.DirEntry, err error) error {
		if err != nil {
			w.logger.Warn("dirwlkr error accessing file", "path", path, "err", err)
			return nil
		}
		if d.IsDir() {
			return nil
		}
		select {
		case <- ctx.Done():
			w.logger.Warn("dirwlkr canceled by context", "path", path)
			return ctx.Err()
		default:
			for _, ch := range w.out {
				w.logger.Info("dirwlkr found file", "path", path)
				ch <- path
			}
		}
		return nil
	}

	err := filepath.WalkDir(w.root, walkFn)
	if err != nil && err != context.Canceled {
		w.logger.Error("dirwlkr finished with error", "err", err)
		return err
	}
	w.logger.Info("dirwlkr completed successfully")
	return nil
}

