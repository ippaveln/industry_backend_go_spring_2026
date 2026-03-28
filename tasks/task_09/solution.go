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
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	out := make([]R, len(in))
	jobs := make(chan int)

	errCh := make(chan error, 1)
	var setErrOnce sync.Once
	setErr := func(err error) {
		if err == nil {
			return
		}
		setErrOnce.Do(func() {
			errCh <- err
			cancel()
		})
	}

	workerCount := workers
	if workerCount > len(in) {
		workerCount = len(in)
	}

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for w := 0; w < workerCount; w++ {
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				case idx, ok := <-jobs:
					if !ok {
						return
					}

					v, err := fn(ctx, in[idx])
					if err != nil {
						setErr(err)
						return
					}

					out[idx] = v
				}
			}
		}()
	}

produce:
	for i := range in {
		select {
		case <-ctx.Done():
			break produce
		case jobs <- i:
		}
	}
	close(jobs)

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
