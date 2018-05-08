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
	sync.WaitGroup
	count int
	ch    chan<- Job
}

// NewPool ...
func NewPool(concurrency int) *Pool {
	ch := make(chan Job)
	p := &Pool{
		count: concurrency,
		ch:    ch,
	}
	proc := makeJobProcessor(ch)
	for i := 0; i < concurrency; i++ {
		p.Add(1)
		go func() {
			defer p.Done()
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
	p.Wait()
}
