package main

import (
	"context"
	"errors"
	"sync"
)

var (
	errWorkerNumber = errors.New("number of workers must be positive")
)

type Req[T any] struct {
	pos int
	val T
}

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) (out []R, err error) {
	if workers < 1 {
		return nil, errWorkerNumber
	}
	if len(in) == 0 {
		return nil, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var once sync.Once
	var wg sync.WaitGroup
	out = make([]R, len(in))
	reqChan := make(chan Req[T])
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var curErr error
			for req := range reqChan {
				select {
				case <-ctx.Done():
					once.Do(func() { err = ctx.Err() })
				default:
					out[req.pos], curErr = fn(ctx, req.val)
					if curErr != nil {
						once.Do(func() { err = curErr })
						cancel()
					}
				}
			}
		}()
	}

	for pos, t := range in {
		select {
		case <-ctx.Done():
			once.Do(func() { err = ctx.Err() })
		case reqChan <- Req[T]{pos, t}:
		}
	}
	close(reqChan)

	wg.Wait()

	if err != nil {
		return nil, err
	}
	return out, nil
}
