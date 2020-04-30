package gpool

import (
	"sync"
)

// Proc ...
type Proc func(<-chan interface{}, chan<- interface{})

// Stage ...
type Stage struct {
	proc       Proc
	concurrent int
	wg         *sync.WaitGroup
}

// NewStage ...
func NewStage(concurrent int, proc Proc) *Stage {
	// For safety
	if concurrent < 1 {
		concurrent = 1
	}
	return &Stage{
		proc:       proc,
		concurrent: concurrent,
		wg:         &sync.WaitGroup{},
	}
}

// Pipeline ...
type Pipeline struct {
	stages []*Stage
}

// New ...
func NewPipeline(stages ...*Stage) (p *Pipeline) {
	p = &Pipeline{make([]*Stage, 0, len(stages))}
	for _, s := range stages {
		p.Add(s)
	}
	return
}

// Add ...
func (p *Pipeline) Add(s *Stage) {
	p.stages = append(p.stages, s)
}

// Start ...
func (p *Pipeline) Start() (chan<- interface{}, <-chan interface{}) {
	var input = make(chan interface{})
	var output = input
	for _, s := range p.stages {
		prev := output
		output = make(chan interface{})
		for i := 0; i < s.concurrent; i++ {
			s.wg.Add(1)
			go func(s *Stage, in <-chan interface{}, out chan<- interface{}) {
				s.proc(in, out)
				s.wg.Done()
			}(s, prev, output)
		}
		go func(s *Stage, out chan<- interface{}) {
			s.wg.Wait()
			close(out)
		}(s, output)
	}
	return input, output
}
