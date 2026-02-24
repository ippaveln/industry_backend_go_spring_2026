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

	if len(in) == 0 {
		return []R{}, nil
	}

	if workers <= 0 {
		return nil, errors.New("workers must be > 0")
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make([]R, len(in))

	type task struct {
		idx   int
		value T
	}

	jobs := make(chan task, min(workers, len(in)))

	var firstErr error
	var once sync.Once
	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}
					res, err := fn(ctx, job.value)
					if err != nil {
						once.Do(func() {
							firstErr = err
							cancel()
						})
						return
					}
					out[job.idx] = res
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
			case jobs <- task{i, v}:
			}
		}
	}()

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return out, nil
}
