package gpool

import (
	"testing"
)

func TestPipeline(t *testing.T) {
	p := NewPipeline(
		NewStage(4, func(in <-chan interface{}, out chan<- interface{}) {
			for v := range in {
				out <- v.(int) * 2
			}
		}),
		NewStage(3, func(in <-chan interface{}, out chan<- interface{}) {
			for v := range in {
				out <- v.(int) + 1
			}
		}),
	)
	in, out := p.Start()
	go func() {
		for i := range make([]struct{}, 10, 10) {
			in <- i
		}
		close(in)
	}()
	total := 0
	for v := range out {
		total += v.(int)
	}
	expected := 9*10 + 10
	if expected != total {
		t.Errorf("Expected %d but got %d", expected, total)
	}
}
