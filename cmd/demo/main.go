package main

import (
	"context"
	"flag"
	"fmt"
	"gopipe/internal"
	"log/slog"
	"os"
	"os/signal"
	"sync"
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

	nodesCount := 3
	errCh := make(chan error, nodesCount)
	var wg sync.WaitGroup

	wg.Add(1)
	go func(){
		defer wg.Done()
		runPipeline(ctx, *dir, *parallelism, nodesCount, logger, errCh) 
	}()

	var pipelineErr error

loop:
	for {
		select {
		case <- ctx.Done():
			logger.Warn("pipeline stopped by context", "reason", ctx.Err())
			break loop
		case err, ok := <- errCh:
			if !ok {
				break loop
			}
			if err != nil {
				logger.Error("pipeline stopped by error", "err", err)
				pipelineErr = err
				stop()
			}
		}
	}
	wg.Wait()
	if pipelineErr == nil {
		logger.Info("pipeline finished successfully")
	}
}



func runPipeline(ctx context.Context, dir string, parallelism, nodesCount int, logger *slog.Logger, errCh chan<- error) {
	walkToMd5 := make(chan string)
	md5ToSink := make(chan internal.FileHash)

	dirWalker := internal.NewDirWalker(dir, 1, 0, logger)
	md5Worker := internal.NewMD5Worker(1, 1, parallelism, logger)
	resultSink := internal.NewResultSink(1, logger)

	dirWalker.Out()[0] = walkToMd5
	md5Worker.In()[0] = walkToMd5
	md5Worker.Out()[0] = md5ToSink
	resultSink.In()[0] = md5ToSink

	var wg sync.WaitGroup
	wg.Add(nodesCount)

	go func() {
		defer wg.Done()
		if err := dirWalker.Run(ctx); err != nil {
			errCh <- fmt.Errorf("dirwalker: %w", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := md5Worker.Run(ctx); err != nil {
			errCh <- fmt.Errorf("md5worker: %w", err)
		}
	}()
	go func() {
		defer wg.Done()
		if err := resultSink.Run(ctx); err != nil {
			errCh <- fmt.Errorf("resultsink: %w", err)
		}
	}()

	wg.Wait()
	close(errCh)
}


	

	

