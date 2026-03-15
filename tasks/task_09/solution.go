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
		return nil, errors.New("workers invalid definition")
	}
	if in == nil || len(in) == 0 {
		return []R{}, nil
	}

	type job struct {
		index int
		value T
	}
	type result struct {
		index int
		value R
		err   error
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan job)
	results := make(chan result)

	factWorkers := min(workers, len(in))
	var wg sync.WaitGroup
	wg.Add(factWorkers)
	for i := 0; i < factWorkers; i++ {
		go func() {
			defer wg.Done()
			for job := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}
				res, err := fn(ctx, job.value)
				select {
				case <-ctx.Done():
					return
				case results <- result{index: job.index, value: res, err: err}:
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
			case jobs <- job{index: i, value: v}:
			}
		}
	}()

	out := make([]R, len(in))
	var err error
	var mu sync.Mutex
	var wgr sync.WaitGroup
	wgr.Add(1)
	go func() {
		defer wgr.Done()
		for result := range results {
			if result.err != nil && err == nil {
				mu.Lock()
				err = result.err
				cancel()
				mu.Unlock()
			}
			out[result.index] = result.value
		}
	}()

	wg.Wait()
	close(results)
	wgr.Wait()

	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return out, nil
	}
}
