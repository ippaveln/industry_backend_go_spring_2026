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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if len(in) == 0 {
		return []R{}, nil
	}

	if workers <= 0 {
		return nil, errors.New("workers must be positive")
	}

	taskCh := make(chan struct {
		idx int
		val T
	})

	results := make([]R, len(in))

	errCh := make(chan error, 1)

	var wg sync.WaitGroup

	for i := 0; i < min(workers, len(in)); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case task, ok := <-taskCh:
					if !ok {
						return
					}

					res, err := fn(ctx, task.val)
					if err != nil {
						select {
						case errCh <- err:
							cancel()
						default:
						}

						return
					}
					results[task.idx] = res
				}
			}
		}()
	}

	for i, v := range in {
		select {
		case <-ctx.Done():
			break
		case taskCh <- struct {
			idx int
			val T
		}{i, v}:
		}
	}

	close(taskCh)

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		return results, nil
	}
}
