package closer

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"
)

var defaultCloser = NewWithLogger(slog.Default(), os.Interrupt, syscall.SIGTERM)

func SetLogger(logger *slog.Logger) {
	defaultCloser.SetLogger(logger)
}

func Add(fns ...func() error) {
	defaultCloser.Add(fns...)
}

func Close() {
	defaultCloser.Close()
	defaultCloser.Wait()
}

type Closer struct {
	mu     sync.Mutex
	once   sync.Once
	done   chan struct{}
	fns    []func() error
	logger *slog.Logger
}

func New(sig ...os.Signal) *Closer {
	return NewWithLogger(slog.Default(), sig...)
}

func NewWithLogger(logger *slog.Logger, sig ...os.Signal) *Closer {
	c := &Closer{
		done:   make(chan struct{}),
		logger: logger,
	}

	if len(sig) > 0 {
		go func() {
			ch := make(chan os.Signal, 1)
			signal.Notify(ch, sig...)
			<-ch
			signal.Stop(ch)
			c.Close()
		}()
	}

	return c
}

func (c *Closer) SetLogger(logger *slog.Logger) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logger = logger
}

func (c *Closer) Add(fns ...func() error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fns = append(c.fns, fns...)
}

func (c *Closer) Wait() {
	<-c.done
}

func (c *Closer) Close() {
	c.once.Do(func() {
		defer close(c.done)

		c.mu.Lock()
		fns := c.fns
		logger := c.logger
		c.fns = nil
		c.mu.Unlock()

		var errs []error

		slices.Reverse(fns)

		for i, fn := range fns {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- fn()
			}()

			select {
			case <-ctx.Done():
				logger.Error("closer timeout", slog.Int("fn_number", i+1), slog.String("error", ctx.Err().Error()))
				errs = append(errs, ctx.Err())
			case ferr := <-done:
				if ferr != nil {
					logger.Error("closer function failed", slog.Int("fn_number", i+1), slog.String("error", ferr.Error()))
					errs = append(errs, ferr)
				} else {
					logger.Info("closer function done", slog.Int("fn_number", i+1))
				}
			}
		}

		if len(errs) > 0 {
			logger.Error("failed to close all functions", slog.String("error", errors.Join(errs...).Error()))
		}
	})
}
