package main

import (
	"context"
	"sync"
)

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) ([]R, error) {
	if len(in) == 0 {
		return []R{}, nil
	}
	if workers <= 0 {
		return nil, context.Canceled
	}
	if workers > len(in) {
		workers = len(in)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type job struct {
		index int
		value T
	}

	out := make([]R, len(in))
	jobs := make(chan job)
	errCh := make(chan error, 1)

	var once sync.Once
	sendErr := func(err error) {
		once.Do(func() {
			errCh <- err
			cancel()
		})
	}

	var wg sync.WaitGroup
	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case j, ok := <-jobs:
					if !ok {
						return
					}

					res, err := fn(ctx, j.value)
					if err != nil {
						sendErr(err)
						return
					}
					out[j.index] = res
				}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for i, v := range in {
			select {
			case <-ctx.Done():
				return
			case jobs <- job{index: i, value: v}:
			}
		}
	}()

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
		if err := ctx.Err(); err != nil && err != context.Canceled {
			return nil, err
		}
		if err := ctx.Err(); err != nil && len(errCh) == 0 {
			return nil, err
		}
		return out, nil
	}
}