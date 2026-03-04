package main

import (
	"context"
	"errors"
	"sync"
)

type Job[T any] struct {
	work T
	id   int
}

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) ([]R, error) {

	if workers <= 0 {
		return nil, errors.New("error")
	}

	if len(in) == 0 {
		return []R{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make([]R, len(in))
	jobs := make(chan Job[T])
	errChan := make(chan error, 1)

	var wg sync.WaitGroup

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				select {
					case <-ctx.Done():
						return
					default:
				}
				answer, err := fn(ctx, job.work)
				if err != nil {
					select {
						case errChan <- err:
						default:
					}
					cancel()
					return
				}
				results[job.id] = answer
			}
		}()
	}

	go func() {
		defer close(jobs)
		for idx, item := range in {
			select {
				case jobs <- Job[T]{work: item, id: idx}:
				case <-ctx.Done():
					return
			}
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	<-done

	select {
	case err := <-errChan:
		return nil, err
	default:
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return results, nil
	}
}
