package main

import (
	"context"
	"errors"
	"sync"
)

var NonPositiveError = errors.New("workers must be > 0")

type job[T any] struct {
	idx int
	val T
}

type res[R any] struct {
	idx int
	val R
	err error
}

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
		return nil, NonPositiveError
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if workers > len(in) {
		workers = len(in)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan job[T])
	results := make(chan res[R], len(in))

	var wg sync.WaitGroup
	wg.Add(workers)

	processTask(ctx, cancel, workers, &wg, jobs, results, fn)
	prepareJobs(ctx, jobs, in)
	closeResults(&wg, results)

	out := make([]R, len(in))
	done := 0
	var firstErr error

	for r := range results {
		if r.err != nil && firstErr == nil {
			firstErr = r.err
			cancel()
			continue
		}
		if r.err == nil {
			out[r.idx] = r.val
			done++
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}
	if done < len(in) {
		return nil, ctx.Err()
	}
	return out, nil
}

func processTask[T any, R any](
	ctx context.Context,
	cancel context.CancelFunc,
	workers int,
	wg *sync.WaitGroup,
	jobs <-chan job[T],
	results chan<- res[R],
	fn func(context.Context, T) (R, error),
) {
	for w := 0; w < workers; w++ {
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

					r, err := fn(ctx, j.val)

					select {
					case results <- res[R]{idx: j.idx, val: r, err: err}:
					case <-ctx.Done():
						return
					}

					if err != nil {
						cancel()
						return
					}
				}
			}
		}()
	}
}

func prepareJobs[T any](
	ctx context.Context,
	jobs chan<- job[T],
	in []T,
) {
	go func() {
		defer close(jobs)
		for i, v := range in {
			select {
			case <-ctx.Done():
				return
			case jobs <- job[T]{idx: i, val: v}:
			}
		}
	}()
}

func closeResults[R any](
	wg *sync.WaitGroup,
	results chan<- res[R],
) {
	go func() {
		wg.Wait()
		close(results)
	}()
}
