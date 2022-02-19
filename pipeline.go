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
	input := make(chan interface{})
	output := input
	for _, s := range p.stages {
		wg := &sync.WaitGroup{}
		prev := output
		output = make(chan interface{})
		for i := 0; i < s.concurrent; i++ {
			wg.Add(1)
			go func(s *Stage, in <-chan interface{}, out chan<- interface{}) {
				s.proc(in, out)
				wg.Done()
			}(s, prev, output)
		}
		go func(s *Stage, out chan<- interface{}) {
			wg.Wait()
			close(out)
		}(s, output)
	}
	return input, output
}

type BranchProc func(<-chan interface{}, []chan<- interface{})

func BranchRun(proc BranchProc, concurrent int, in <-chan interface{}, outs []chan<- interface{}) {
	var wg sync.WaitGroup
	for i := 0; i < concurrent; i++ {
		wg.Add(1)
		go func(in <-chan interface{}, outs []chan<- interface{}) {
			proc(in, outs)
			wg.Done()
		}(in, outs)
	}
	go func(outs []chan<- interface{}) {
		wg.Wait()
		for _, out := range outs {
			close(out)
		}
	}(outs)
}

func Branch(in <-chan interface{}, outs []chan<- interface{}) {
	BranchRun(func(in <-chan interface{}, outs []chan<- interface{}) {
		for v := range in {
			for _, out := range outs {
				out <- v
			}
		}
	}, 1, in, outs)
}
