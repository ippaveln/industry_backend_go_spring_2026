package main

import (
	"context"
	"errors"
	"sync"
)

type Job[T any] struct {
	index int
	value T
}

type Res[R any] struct {
	index int
	value R
	err   error
}

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) ([]R, error) {

	if workers <= 0 {
		return nil, errors.New("workers must be positive")
	}
	if len(in) == 0 {
		return []R{}, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan Job[T])
	results := make(chan Res[R])
	out := make([]R, len(in))
	var wg sync.WaitGroup

	if workers > len(in) {
		workers = len(in)
	}
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for j := range jobs {
				r, err := fn(ctx, j.value)
				results <- Res[R]{index: j.index, value: r, err: err}
			}
		}()
	}

	go func() {
		defer close(jobs)
		for i, v := range in {
			select {
			case <-ctx.Done():
				return
			case jobs <- Job[T]{index: i, value: v}:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var firsterr error
	for r := range results {
		if r.err != nil && firsterr == nil {
			cancel()
			firsterr = r.err
		} else if r.err == nil {
			out[r.index] = r.value
		}
	}

	return out, firsterr
}
