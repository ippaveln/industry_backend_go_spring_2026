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

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if len(in) == 0 {
		return []R{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type workItem struct {
		idx int
		val T
	}

	type result struct {
		idx int
		val R
		err error
	}

	workCh := make(chan workItem)
	resultCh := make(chan result, len(in))

	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range workCh {
				r, err := fn(ctx, item.val)
				resultCh <- result{idx: item.idx, val: r, err: err}
			}
		}()
	}

	go func() {
		defer close(workCh)
		for i, v := range in {
			select {
			case <-ctx.Done():
				return
			case workCh <- workItem{idx: i, val: v}:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	out := make([]R, len(in))
	var firstErr error

	for r := range resultCh {
		if r.err != nil && firstErr == nil {
			firstErr = r.err
			cancel()
		}
		if r.err == nil {
			out[r.idx] = r.val
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}
	return out, nil
}
