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
		return nil, errors.New("invalid workers count")
	}
	if len(in) == 0 {
		return make([]R, 0), nil
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make([]R, len(in))

	type job struct {
		index int
		value T
	}

	actualWorkers := workers
	if actualWorkers > len(in) {
		actualWorkers = len(in)
	}

	jobs := make(chan job, actualWorkers)

	type result struct {
		index int
		value R
		err   error
	}
	results := make(chan result, actualWorkers)

	// Sender goroutine
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

	var wg sync.WaitGroup
	for i := 0; i < actualWorkers; i++ {
		wg.Go(func() {
			for j := range jobs {
				if err := ctx.Err(); err != nil {
					results <- result{index: j.index, err: err}
					continue
				}

				res, err := fn(ctx, j.value)
				results <- result{index: j.index, value: res, err: err}

				if err != nil {
					cancel()
				}
			}
		})
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var firstErr error
	var processedCount int
	for r := range results {
		processedCount++
		if r.err != nil {
			if firstErr == nil {
				firstErr = r.err
				cancel()
			}
		} else {
			out[r.index] = r.value
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	if processedCount < len(in) {
		return nil, ctx.Err()
	}

	return out, nil
}
