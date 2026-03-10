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
	// краевые случаи
	if workers <= 0 {
		return nil, errors.New("workers <= 0")
	}

	if workers > len(in) {
		workers = len(in)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(in) == 0 {
		return []R{}, nil
	}

	type job struct {
		idx int
		val T
	}

	type result struct {
		idx int
		val R
		err error
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan job)
	results := make(chan result)

	var wg sync.WaitGroup
	wg.Add(workers)

	// workers
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			for j := range jobs {
				v, err := fn(ctx, j.val)

				select {
				case results <- result{idx: j.idx, val: v, err: err}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// ждем воркеров,закрываем канал
	go func() {
		wg.Wait()
		close(results)
	}()

	// producer
	go func() {
		defer close(jobs)

		for i, v := range in {
			select {
			case jobs <- job{idx: i, val: v}:
			case <-ctx.Done():
				return
			}
		}
	}()

	out := make([]R, len(in))
	var firstErr error

	for res := range results {
		if res.err != nil && firstErr == nil {
			firstErr = res.err
			cancel()
			continue
		}

		if firstErr == nil {
			out[res.idx] = res.val
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return out, nil
}
