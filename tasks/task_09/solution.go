package main

import (
	"context"
	"errors"
	"sync"
)

type item[T any] struct {
	index int
	value T
}

func reader[T any](ctx context.Context, in []T, ch chan<- item[T], wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(ch)
	
	for i, val := range in {
		select {
		case <- ctx.Done():
			return
		case ch <-item[T]{index: i, value: val}:
		}
	}
}

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

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Честно говоря, не могу решить стоит ли тут использовать буферизированный канал
	// Реализованная ф-я reader() больше похожа на генератор, 
	// а в моем понимании генератору не нужен буфер, буду рад, если поправите меня
	// Пока оставил небуферизированный
	ch := make(chan item[T]) 
	out := make(chan item[R], len(in))
	var errorValue error
	var mtx sync.Mutex
	var errOnce sync.Once

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := &sync.WaitGroup{}


	setFirstErr := func(err error) {
		errOnce.Do(func() {
			mtx.Lock()
			errorValue = err
			mtx.Unlock()
			cancel()
		})
	}

	wg.Add(1)
	go reader(ctx, in, ch, wg)

	for range workers {
		wg.Add(1)
		go func () {
			defer wg.Done()

			for v := range ch {
				select {
				case <-ctx.Done():
					setFirstErr(ctx.Err())
					return
				default:
					res, err := fn(ctx, v.value)
					if err != nil {
						setFirstErr(err)
						return
					}

					select {
					case <-ctx.Done():
						setFirstErr(ctx.Err())
						return
					case out <- item[R]{index: v.index, value: res}:
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	results := make([]R, len(in))
	done := make(chan struct{})

	go func() {
		defer close(done)
		for it := range out {
			results[it.index] = it.value
		}
	}()

	select {
	case <-done:
		if errorValue != nil {
			return results, errorValue
		}
		return results, nil
	case <-ctx.Done():
		<-done
		if errorValue != nil {
			return results, errorValue
		}
		return results, ctx.Err()
	}
}