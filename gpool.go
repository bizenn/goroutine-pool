package gpool

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// Job ...
type Job interface {
	Do() error
	LogProperties() logrus.Fields
	SetError(error)
	GetError() error
	// Assumed to embed sync.WaitGroup in
	Add(int)
	Done()
	Wait()
}

// BaseJob ...
type BaseJob struct {
	Job
	sync.WaitGroup
	err error
}

// SetError ...
func (j *BaseJob) SetError(err error) {
	j.err = err
}

// GetError ...
func (j *BaseJob) GetError() error {
	return j.err
}

// Add ...
func (j *BaseJob) Add(delta int) {
	j.WaitGroup.Add(delta)
}

// Done ...
func (j *BaseJob) Done() {
	j.WaitGroup.Done()
}

// Wait ...
func (j *BaseJob) Wait() {
	j.WaitGroup.Wait()
}

// ProcJob ...
type ProcJob struct {
	BaseJob
	proc func() error
}

// NewProcJob ...
func NewProcJob(proc func() error) *ProcJob {
	j := &ProcJob{
		proc: proc,
	}
	j.Add(1)
	return j
}

// LogProperties ...
func (j *ProcJob) LogProperties() logrus.Fields {
	return logrus.Fields{
		"Name": "ProcJob",
	}
}

// Do ...
func (j *ProcJob) Do() error {
	defer j.Done()
	return j.proc()
}

func makeJobProcessor(ch <-chan Job) func() {
	return func() {
		for {
			select {
			case job, ok := <-ch:
				if ok {
					job.Add(1)
					proc := func() {
						defer job.Done()
						job.SetError(job.Do())
					}
					proc()
					continue
				}
				break
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
