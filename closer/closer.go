package closer

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const shutdownTimeout = 5 * time.Second

type closeFn struct {
	fn     func(context.Context) error
	source string
}

type Closer struct {
	mu     sync.Mutex
	once   sync.Once
	done   chan struct{}
	funcs  []closeFn
	logger *slog.Logger

	rootCtx    context.Context
	rootCancel context.CancelFunc

	grp    *errgroup.Group
	grpCtx context.Context

	firstErrOnce sync.Once
	firstErr     error

	isGlobal bool
}

type Task struct {
	c         *Closer
	startedCh chan struct{}
}

var defaultCloser = func() *Closer {
	c := New()
	c.isGlobal = true
	return c
}()

func SetLogger(l *slog.Logger)                                { defaultCloser.SetLogger(l) }
func Context() context.Context                                { return defaultCloser.Context() }
func ToClose(fns ...func(context.Context) error)              { defaultCloser.ToClose(fns...) }
func ToCloseNamed(name string, f func(context.Context) error) { defaultCloser.ToCloseNamed(name, f) }
func Go(fn func(context.Context) error) *Task                 { return defaultCloser.Go(fn) }
func Close(ctx context.Context) error                         { return defaultCloser.Close(ctx) }
func Wait() error                                             { return defaultCloser.Wait() }

func New(signals ...os.Signal) *Closer {
	return NewWithLogger(slog.Default(), signals...)
}

func NewWithLogger(logger *slog.Logger, signals ...os.Signal) *Closer {
	if logger == nil {
		logger = slog.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Closer{
		done:       make(chan struct{}),
		logger:     logger,
		rootCtx:    ctx,
		rootCancel: cancel,
	}
	c.grp, c.grpCtx = errgroup.WithContext(c.rootCtx)

	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}

	go c.handleSignals(signals...)

	return c
}

func (c *Closer) Context() context.Context { return c.rootCtx }

func (c *Closer) SetLogger(l *slog.Logger) {
	c.logger = l
}

func (c *Closer) ToClose(fns ...func(context.Context) error) {
	skip := 2
	if c.isGlobal {
		skip = 3
	}
	src := callerName(skip)

	c.mu.Lock()
	for _, f := range fns {
		c.funcs = append(c.funcs, closeFn{fn: f, source: src})
	}
	c.mu.Unlock()
}

func (c *Closer) ToCloseNamed(name string, f func(context.Context) error) {
	skip := 2
	if c.isGlobal {
		skip = 3
	}
	src := callerName(skip)

	wrapped := func(ctx context.Context) error {
		start := time.Now()
		c.logger.Info("closing dependency start", slog.String("name", name), slog.String("source", src))
		err := f(ctx)
		d := time.Since(start)
		if err != nil {
			c.logger.Error("closing dependency failed", slog.String("name", name), slog.String("source", src), slog.Duration("duration", d), slog.String("error", err.Error()))
			return err
		}
		c.logger.Info("closing dependency done", slog.String("name", name), slog.String("source", src), slog.Duration("duration", d))
		return nil
	}

	c.mu.Lock()
	c.funcs = append(c.funcs, closeFn{fn: wrapped, source: src})
	c.mu.Unlock()
}

func (c *Closer) Go(fn func(context.Context) error) *Task {
	t := &Task{
		c:         c,
		startedCh: make(chan struct{}),
	}
	c.grp.Go(func() error {
		close(t.startedCh)
		err := fn(c.grpCtx)
		if err != nil {
			c.setFirstErr(err)
			c.initiateShutdown()
		}
		return err
	})
	return t
}

func (t *Task) With(fn func(context.Context) error) *Task {
	t.c.grp.Go(func() error {
		<-t.startedCh
		err := fn(t.c.grpCtx)
		if err != nil {
			t.c.setFirstErr(err)
			t.c.initiateShutdown()
		}
		return err
	})
	return t
}

func (c *Closer) Close(ctx context.Context) error {
	var result error

	c.once.Do(func() {
		defer close(c.done)

		c.rootCancel()

		c.mu.Lock()
		funcs := slices.Clone(c.funcs)
		c.funcs = nil
		c.mu.Unlock()

		if len(funcs) == 0 {
			c.logger.Info("no resources to close")
			return
		}

		c.logger.Info("starting graceful shutdown", slog.Int("count", len(funcs)))
		slices.Reverse(funcs)

		for i, item := range funcs {
			if err := ctx.Err(); err != nil {
				c.logger.Warn("shutdown context canceled", slog.String("error", err.Error()))
				if result == nil {
					result = err
				}
				break
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						c.logger.Error("panic recovered in closer", slog.Int("index", i), slog.Any("panic", r), slog.String("source", item.source))
						if result == nil {
							result = errors.New("panic recovered in closer")
						}
					}
				}()

				start := time.Now()
				if err := item.fn(ctx); err != nil {
					c.logger.Error("closer returned error", slog.Duration("duration", time.Since(start)), slog.String("error", err.Error()), slog.String("source", item.source))
					if result == nil {
						result = err
					}
					return
				}

				c.logger.Info("closer completed", slog.Duration("duration", time.Since(start)), slog.String("source", item.source))
			}()
		}

		if result == nil {
			c.logger.Info("graceful shutdown completed successfully")
		} else {
			c.logger.Error("graceful shutdown completed with errors", slog.String("error", result.Error()))
		}
	})

	return result
}

func (c *Closer) Wait() error {
	grpDone := make(chan error, 1)
	go func() { grpDone <- c.grp.Wait() }()

	select {
	case err := <-grpDone:
		c.setFirstErr(err)
		if !c.isClosed() {
			c.initiateShutdown()
		}
	case <-c.done:
	}

	<-c.done
	return c.firstErr
}

func (c *Closer) handleSignals(signals ...os.Signal) {
	ch := make(chan os.Signal, len(signals))
	signal.Notify(ch, signals...)
	defer signal.Stop(ch)

	select {
	case <-ch:
		c.logger.Info("signal received, initiating shutdown")
		c.initiateShutdown()
	case <-c.done:
	}
}

func (c *Closer) initiateShutdown() {
	sdCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	_ = c.Close(sdCtx)
}

func (c *Closer) setFirstErr(err error) {
	if err == nil {
		return
	}
	c.firstErrOnce.Do(func() { c.firstErr = err })
}

func (c *Closer) isClosed() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

func callerName(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	full := fn.Name()
	if idx := strings.LastIndex(full, "/"); idx != -1 {
		return full[idx+1:]
	}
	return full
}
