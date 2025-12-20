package diut

import (
	"context"
	"runtime"
	"strconv"
	"sync"

	"golang.org/x/sync/singleflight"
)

var (
	onceStore     sync.Map
	singleRequest singleflight.Group
)

type onceEntry struct {
	val any
}

func callerKey(skip int) string {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown"
	}

	funcName := runtime.FuncForPC(pc).Name()
	return file + ":" + strconv.Itoa(line) + ":" + funcName
}

func Once[T any](ctx context.Context, f func(context.Context) T) T {
	key := callerKey(1)

	if val, ok := onceStore.Load(key); ok {
		return val.(*onceEntry).val.(T)
	}

	result, _, _ := singleRequest.Do(key, func() (interface{}, error) {
		val := f(ctx)
		onceStore.Store(key, &onceEntry{val: val})
		return val, nil
	})

	return result.(T)
}
