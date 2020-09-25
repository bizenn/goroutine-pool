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
		NewStage(1, func(in <-chan interface{}, out chan<- interface{}) {
			var total int
			for v := range in {
				total += v.(int)
			}
			out <- total
		}),
	)
	in, out := p.Start()
	go func() {
		for i := 0; i < 10; i++ {
			in <- i
		}
		close(in)
	}()
	expected := (1 + 19) * 10 / 2
	total := <-out
	if expected != total.(int) {
		t.Errorf("Expected %d but got %d", expected, total.(int))
	}
}
