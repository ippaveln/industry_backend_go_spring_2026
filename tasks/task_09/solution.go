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

	if len(in) == 0 {
		return []R{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type task struct {
		index int
		value T
	}

	type result struct {
		index int
		value R
		err   error
	}

	tasks := make(chan task, len(in))
	results := make(chan result, len(in))

	for i, v := range in {
		tasks <- task{index: i, value: v}
	}
	close(tasks)

	var wg sync.WaitGroup
	numWorkers := workers
	if numWorkers > len(in) {
		numWorkers = len(in)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for task := range tasks {
				select {
				case <-ctx.Done():
					return
				default:
				}

				val, err := fn(ctx, task.value)

				select {
				case results <- result{index: task.index, value: val, err: err}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	output := make([]R, len(in))
	var firstErr error
	var errMutex sync.Mutex
	resultsReceived := 0

	for res := range results {
		if res.err != nil {
			errMutex.Lock()
			if firstErr == nil {
				firstErr = res.err
				cancel()
			}
			errMutex.Unlock()
		}

		output[res.index] = res.value
		resultsReceived++

		if resultsReceived == len(in) {
			break
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return output, nil
	}
}
