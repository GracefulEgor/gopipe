package internal

import (
	"context"
	"log/slog"
	"sync"
)


type ResultSink struct {
	in []<-chan FileHash
	logger *slog.Logger
}

func NewResultSink(numIn int, logger *slog.Logger) *ResultSink {
	in := make([]<-chan FileHash, numIn)
	return &ResultSink{in: in, logger: logger}
}

func (s *ResultSink) In() []<-chan FileHash {return s.in}

func (s *ResultSink) Out() []chan<- FileHash {return nil}


func (s *ResultSink) Run(ctx context.Context) error {
	s.logger.Info("rsltSink started")
	var wg sync.WaitGroup

	for _, ch := range s.in {
		wg.Add(1)
		go func(ch <-chan FileHash) {
			defer wg.Done()
			for res := range ch {
				if res.Err != nil {
					s.logger.Error("rsltSink error", "path", res.Path, "err", res.Err)
				} else {
					s.logger.Info("rsltSink got result", "path", res.Path, "hash", res.Hash)
				}
				select {
				case <- ctx.Done():
					s.logger.Warn("rsltSink canceled by context", "path", res.Path)
					return 
				default:
				}
			}
		}(ch)

	}
	wg.Wait()
	s.logger.Info("rsltSink finished")
	return nil
}