package main

import (
	"context"
	"flag"
	"fmt"
	"gopipe/internal"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	dir := flag.String("dir", ".", "dir for walking")
	parallelism := flag.Int("p", 10, "level for parallelism for MD5")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	logger.Info("pipeline started", "dir", *dir, "parallelism", *parallelism)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 3)
	go runPipeline(ctx, *dir, *parallelism, logger, errCh) 

	select {
	case <- ctx.Done():
		logger.Warn("pipeline stopped by context", "reason", ctx.Err())
	case err := <- errCh:
		if err != nil {
			logger.Error("pipeline stopped by error", "err", err)
			stop()
		} else {
			logger.Info("pipeline finished successfully")
		}
	}
}


func runPipeline(ctx context.Context, dir string, parallelism int, logger *slog.Logger, errCh chan<- error) {
	walkToMd5 := make(chan string)
	md5ToSink := make(chan internal.FileHash)

	dirWalker := internal.NewDirWalker(dir, 1, 0, logger)
	md5Worker := internal.NewMD5Worker(1, 1, parallelism, logger)
	resultSink := internal.NewResultSink(1, logger)

	dirWalker.Out()[0] = walkToMd5
	md5Worker.In()[0] = walkToMd5
	md5Worker.Out()[0] = md5ToSink
	resultSink.In()[0] = md5ToSink

	go func() {
		if err := dirWalker.Run(ctx); err != nil {
			errCh <- fmt.Errorf("dirwalker: %w", err)
		}
	}()
	go func() {
		if err := md5Worker.Run(ctx); err != nil {
			errCh <- fmt.Errorf("md5worker: %w", err)
		}
	}()
	go func() {
		if err := resultSink.Run(ctx); err != nil {
			errCh <- fmt.Errorf("resultsink: %w", err)
		} else {
			errCh <- nil 
		}
	}()
}


	

	

