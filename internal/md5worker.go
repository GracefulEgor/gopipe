package internal

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"log/slog"
	"os"
	"sync"
)

type FileHash struct {
	Path string
	Hash string
	Err error
}

type MD5Worker struct {
	in []<-chan string
	out []chan<- FileHash
	parallelism int
	logger *slog.Logger
}

func NewMD5Worker(numIn, numOut, parallelism int, logger *slog.Logger) *MD5Worker {
	in := make([]<-chan string, numIn)
	out := make([]chan<- FileHash, numOut)
	return &MD5Worker{
		in: in,
		out: out,
		parallelism: parallelism,
		logger: logger,
	}
}

func (w *MD5Worker) In() []<-chan string {return w.in}

func (w *MD5Worker) Out() []chan<- FileHash {return w.out}


func (w *MD5Worker) Run(ctx context.Context) error {
	w.logger.Info("md5wrkr started", "parallelism", w.parallelism)
	defer func() {
		for _, ch := range w.out {
			close(ch)
		}
		w.logger.Info("md5wrkr finished")
	}()

	var wg sync.WaitGroup
	jobs := make(chan string)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(jobs)
		for _, inCh := range w.in {
			for path := range inCh {
				select {
				case <-ctx.Done():
					w.logger.Warn("md5wrkr canceled by context (input)", "path", path)
					return
				case jobs <- path:
				}
			}
		}
	}()

	wg.Add(w.parallelism)
	for i := 0; i < w.parallelism; i++ { 
		go func(workerID int) {
			defer wg.Done()
			for path := range jobs {
				w.logger.Debug("md5wrkr processing file", "worker", workerID, "path", path)
				hash, err := computeMD5(path)
				if err != nil {
					w.logger.Error("failed to compute MD5", "path", path, "err", err)
				}
				res := FileHash{Path: path, Hash: hash, Err: err}
				for _, outCh := range w.out {
					select {
					case <- ctx.Done():
						w.logger.Warn("md5wrkr canceled by context (output)", "path", path)
						return
					case outCh <- res:
					}
				}
			}
		}(i)
	}

	wg.Wait()
	w.logger.Info("md5wrkr all workers finished")
	return nil

}


func computeMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := md5.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
