package main

import (
	"context"
	"errors"
	"sync"
)

type job[T any] struct {
	id   int
	data T
}

var ErrInvalidWorkers = errors.New("invalid number of workers")

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) ([]R, error) {
	out := make([]R, len(in))

	if workers <= 0 {
		return out, ErrInvalidWorkers
	}
	if len(in) == 0 {
		return out, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(workers)
	jobs := make(chan job[T], workers)

	go func() {
		defer close(jobs)
		for i, v := range in {
			job := job[T]{
				id:   i,
				data: v,
			}
			select {
			case <-ctx.Done():
				return
			case jobs <- job:
			}
		}
	}()

	var fnErr error
	var mu sync.Mutex
	for range workers {
		go func() {
			defer wg.Done()
			for job := range jobs {
				if ctx.Err() != nil {
					return
				}

				res, err := fn(ctx, job.data)
				if err != nil {
					mu.Lock()
					if fnErr == nil {
						fnErr = err
					}
					mu.Unlock()
					cancel()
					return
				}
				out[job.id] = res
			}
		}()
	}

	wg.Wait()
	if fnErr != nil {
		return nil, fnErr
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return out, nil
}
