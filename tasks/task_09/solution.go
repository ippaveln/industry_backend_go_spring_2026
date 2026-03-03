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
		return nil, nil
	}

	if workers <= 0 {
		return nil, errors.New("negative workers count")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type task struct {
		i int
		v T
	}
	jobs := make(chan task, workers)
	out := make([]R, len(in))
	var wg sync.WaitGroup
	var firstErr error
	var once sync.Once

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-jobs:
					if !ok {
						return
					}

					if ctx.Err() != nil {
						return
					}

					res, err := fn(ctx, task.v)
					if err != nil {
						once.Do(func() {
							firstErr = err
							cancel()
						})
						return
					}

					out[task.i] = res
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer close(jobs)
		defer wg.Done()
		for i, v := range in {
			select {
			case <-ctx.Done():
				return
			case jobs <- task{i: i, v: v}:
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
