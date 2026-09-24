package worker

import (
	"context"
	"fmt"
	"sync"
)

type Job[T any] struct {
	ID    int
	Value T
}

type Result[T any, U any] struct {
	Job   Job[T]
	Value U
	Err   error
}

type ProcessFunc[T any, U any] func(context.Context, T) (U, error)

func tryRun[T any, U any](ctx context.Context, fn ProcessFunc[T, U], v T) (u U, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("worker panic: %v", r)
		}
	}()
	return fn(ctx, v)
}

func worker[T any, U any](ctx context.Context, process ProcessFunc[T, U], jobs <-chan Job[T], results chan<- Result[T, U]) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			value, err := tryRun(ctx, process, job.Value)

			select {
			case results <- Result[T, U]{Job: job, Value: value, Err: err}:
			case <-ctx.Done():
				return
			}
		}
	}
}

func Run[T any, U any](ctx context.Context, size int, jobs []T, fn ProcessFunc[T, U]) []Result[T, U] {
	results := make([]Result[T, U], len(jobs))
	if len(jobs) == 0 {
		return results
	}
	if size < 1 {
		size = 1
	}

	jobCh := make(chan Job[T], len(jobs))
	resCh := make(chan Result[T, U])

	var wg sync.WaitGroup
	for range size {
		wg.Go(func() {
			worker(ctx, fn, jobCh, resCh)
		})
	}
	go func() {
		wg.Wait()
		close(resCh)
	}()

	for i, j := range jobs {
		jobCh <- Job[T]{ID: i, Value: j}
	}
	close(jobCh)

	for res := range resCh {
		results[res.Job.ID] = res
	}

	return results
}
