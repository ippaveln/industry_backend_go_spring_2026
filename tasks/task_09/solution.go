package main

import (
	"context"
	"errors"
	"sync"
)

var ErrNoWorkers = errors.New("workers must be positive")

type item[T any] struct {
	idx int
	t   T
}

func generate[T any](in []item[T]) <-chan item[T] {
	out := make(chan item[T], len(in))
	go func() {
		for _, v := range in {
			out <- v
		}
		close(out)
	}()

	return out
}

func createItmes[T any](in []T) []item[T] {
	out := make([]item[T], 0, len(in))
	for i, t := range in {
		out = append(out, item[T]{idx: i, t: t})
	}
	return out
}

func ParallelMap[T any, R any](
	ctx context.Context,
	workers int,
	in []T,
	fn func(context.Context, T) (R, error),
) ([]R, error) {
	if workers <= 0 {
		return nil, ErrNoWorkers
	}
	items := createItmes(in)
	tCh := generate(items)
	results := make([]R, len(in))

	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)

	setErr := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel()
		})
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx2.Done():
					return
				case item, ok := <-tCh:
					if !ok {
						return
					}
					r, err := fn(ctx2, item.t)
					if err != nil {
						setErr(err)
						return 
					}
					results[item.idx] = r
				}
			}
		}()
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return results, nil
}

