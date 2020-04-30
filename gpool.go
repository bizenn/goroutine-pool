package gpool

// Job ...
type Job interface {
	Do()
}

// Pool ...
type Pool struct {
	pipeline *Pipeline
	ch       chan Job
	in       chan<- interface{}
	out      <-chan interface{}
}

// NewPool ...
func NewPool(concurrency int) (p *Pool) {
	p = &Pool{
		pipeline: NewPipeline(NewStage(concurrency, func(in <-chan interface{}, out chan<- interface{}) {
			for v := range in {
				job := v.(Job)
				job.Do()
			}
		})),
		ch: make(chan Job),
	}
	p.in, p.out = p.pipeline.Start()
	// Just convert chan<- Job to chan<- interface{}
	go func() {
		for job := range p.ch {
			p.in <- job
		}
		close(p.in)
	}()
	return p
}

// Channel ...
func (p *Pool) Channel() chan<- Job {
	return p.ch
}

// Count ...
func (p *Pool) Count() int {
	return p.pipeline.stages[0].concurrent
}

// Shutdown ...
func (p *Pool) Shutdown() {
	close(p.Channel())
}

// Wait ...
func (p *Pool) Wait() {
	<-p.out
}
