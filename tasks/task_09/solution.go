package main

import (
	"context"
	"errors"
	"sync"
)

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) ([]R, error) {

	if workers <= 0 {
		return nil, errors.New("workers must be positive")
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if len(in) == 0 {
		return []R{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type job struct {
		idx int
		val T
	}

	jobs := make(chan job)
	out := make([]R, len(in))

	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	setError := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel()
		})
	}

	worker := func() {
		defer wg.Done()

		for j := range jobs {
			res, err := fn(ctx, j.val)
			if err != nil {
				setError(err)
				return
			}
			out[j.idx] = res
		}
	}

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}

loop:
	for i, v := range in {
		select {
		case <-ctx.Done():
			break loop
		case jobs <- job{i, v}:
		}
	}

	close(jobs)
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return out, nil
}
