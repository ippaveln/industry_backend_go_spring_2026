package main

import (
	"context"
	"sync"
)

type Task[T any] struct {
	Index int
	Value T
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
		return nil, context.Canceled
	}

	if workers > len(in) {
		workers = len(in)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	results := make([]R, len(in))
	tasks := make(chan Task[T], len(in))

	var wg sync.WaitGroup

	var firstErr error
	var errMu sync.Mutex

	for i := 0; i < workers; i++ {
		wg.Go(func() {
			for task := range tasks {
				select {
				case <-ctx.Done():
					return
				default:
				}

				res, err := fn(ctx, task.Value)
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()

					cancel()
					return
				}
				results[task.Index] = res
			}
		})
	}

	for i, v := range in {
		tasks <- Task[T]{Index: i, Value: v}
	}
	close(tasks)

	wg.Wait()

	return results, firstErr
}
