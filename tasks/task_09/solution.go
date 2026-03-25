package main

import (
	"context"
	"fmt"
	"sync"
)

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) ([]R, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if workers <= 0 {
		return nil, fmt.Errorf("Non-positive workers count (%d)", workers)
	}
	if len(in) == 0 {
		return []R{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type task struct {
		idx int
		val T
	}
	tasks := make(chan task)

	results := make([]R, len(in))
	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case t, ok := <-tasks:
					if !ok {
						return
					}
					res, err := fn(ctx, t.val)
					if err != nil {
						select {
						case errCh <- err:
						default:
						}
						cancel()
						return
					}
					results[t.idx] = res
				}
			}
		}()
	}

	for i, v := range in {
		select {
		case <-ctx.Done():
			break
		case tasks <- task{idx: i, val: v}:
		}
	}
	close(tasks)

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
		return results, nil
	}
}
