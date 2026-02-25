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
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	if workers <= 0 {
		cancel()
		return nil, ctx.Err()
	}
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(workers)
	out := make([]R, len(in))
	jobs := make(chan int)
	var firstErr error
	var once sync.Once

	worker := func() {
		defer wg.Done()
		for j := range jobs {
			if ctx.Err() != nil {
				return
			}
			res, err := fn(ctx, in[j])
			if err != nil {
				once.Do(func() {
					firstErr = err
					cancel()
				})
			}
			out[j] = res
		}
	}

	for range workers {
		go worker()
	}

	for i := range in {
		select {
		case <-ctx.Done():
			break
		case jobs <- i:
		}
	}
	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return out, nil
}
