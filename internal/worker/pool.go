package worker

import (
	"context"
	"sync"
)

type Job struct {
	PostID int
}

type WorkerPool struct {
	workers int
	jobs    chan Job
	wg      sync.WaitGroup
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewWorkerPool(ctx context.Context, workers int, jobHandler func(Job)) *WorkerPool {
	ctx, cancel := context.WithCancel(ctx)
	pool := &WorkerPool{
		workers: workers,
		jobs:    make(chan Job, workers*2),
		ctx:     ctx,
		cancel:  cancel,
	}

	for i := 0; i < workers; i++ {
		pool.wg.Add(1)
		go pool.worker(jobHandler)
	}

	return pool
}

func (p *WorkerPool) worker(jobHandler func(Job)) {
	defer p.wg.Done()
	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			jobHandler(job)
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *WorkerPool) Submit(job Job) {
	select {
	case p.jobs <- job:
	case <-p.ctx.Done():
		return
	}
}

func (p *WorkerPool) Stop() {
	p.cancel()
	close(p.jobs)
	p.wg.Wait()
}
