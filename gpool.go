package gpool

import (
	"sync"
)

// Job ...
type Job interface {
	Do() error
}

// ProcJob ...
type ProcJob struct {
	Job
	*sync.WaitGroup
	err  error
	proc func() error
}

// NewProcJob ...
func NewProcJob(proc func() error) *ProcJob {
	j := &ProcJob{
		proc:      proc,
		WaitGroup: new(sync.WaitGroup),
	}
	j.Add(1)
	return j
}

// Do ...
func (j *ProcJob) Do() error {
	defer j.Done()
	j.err = j.proc()
	return j.err
}

// GetError ...
func (j *ProcJob) GetError() error {
	j.Wait()
	return j.err
}

func makeJobProcessor(ch <-chan Job) func() {
	return func() {
		for {
			select {
			case job, ok := <-ch:
				if ok {
					proc := func() {
						_ = job.Do()
					}
					proc()
					continue
				}
			}
			break
		}
	}
}

// Pool ...
type Pool struct {
	wg    *sync.WaitGroup
	count int
	ch    chan<- Job
}

// NewPool ...
func NewPool(concurrency int) *Pool {
	if concurrency < 1 {
		concurrency = 1
	}
	ch := make(chan Job)
	p := &Pool{
		wg:    new(sync.WaitGroup),
		count: concurrency,
		ch:    ch,
	}
	proc := makeJobProcessor(ch)
	for i := 0; i < concurrency; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			proc()
		}()
	}
	return p
}

// Channel ...
func (p *Pool) Channel() chan<- Job {
	return p.ch
}

// Count ...
func (p *Pool) Count() int {
	return p.count
}

// Shutdown ...
func (p *Pool) Shutdown() {
	close(p.ch)
	p.wg.Wait()
}
