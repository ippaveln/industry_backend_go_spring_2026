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
		return nil, errors.New("number of workers must be positive")
	}

	if len(in) == 0 {
		return []R{}, nil
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make([]R, len(in))

	tasks := make(chan int, len(in))

	var (
		errOnce  sync.Once
		firstErr error
	)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range tasks {
				if ctx.Err() != nil {
					return
				}

				result, err := fn(ctx, in[idx])

				if err != nil {
					errOnce.Do(func() {
						firstErr = err
						cancel()
					})
					return
				}

				results[idx] = result
			}
		}()
	}

	go func() {
		for i := range in {
			select {
			case <-ctx.Done():
				close(tasks)
				return
			case tasks <- i:
			}
		}
		close(tasks)
	}()

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
