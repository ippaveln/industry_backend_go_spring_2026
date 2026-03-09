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
	if in == nil || len(in) == 0 {
		return nil, nil
	}

	if workers < 1 {
		return nil, errors.New("workers argument must be positive")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var (
		wg       sync.WaitGroup
		firstErr error
		once     sync.Once
	)

	wg.Add(workers)

	result := make([]R, len(in))
	indexesChan := generateIndicesChannel(len(in))

	ctx, cancel := context.WithCancel(ctx)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			for j := range indexesChan {
				var err error

				select {
				case <-ctx.Done():
					once.Do(func() {
						firstErr = ctx.Err()
					})
				default:
					result[j], err = fn(ctx, in[j])

					if err != nil {
						once.Do(func() {
							firstErr = err
						})
						cancel()
					}
				}
			}
		}()
	}

	wg.Wait()
	cancel()

	return result, firstErr
}

func generateIndicesChannel(inputLen int) chan int {
	indexesChan := make(chan int, inputLen)

	for el := 0; el < inputLen; el++ {
		indexesChan <- el
	}
	close(indexesChan)

	return indexesChan
}
