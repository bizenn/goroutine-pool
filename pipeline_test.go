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

	_, _ = p.Start()
	in, out = p.Start()
	go func() {
		for i := 0; i < 10; i++ {
			in <- i
		}
		close(in)
	}()
	expected = (1 + 19) * 10 / 2
	total = <-out
	if expected != total.(int) {
		t.Errorf("Expected %d but got %d", expected, total.(int))
	}
}

func TestBranch(t *testing.T) {
	p1 := NewPipeline(
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
	p2 := NewPipeline(
		NewStage(4, func(in <-chan interface{}, out chan<- interface{}) {
			for v := range in {
				out <- v.(int) * 3
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

	in1, out1 := p1.Start()
	in2, out2 := p2.Start()

	in := make(chan interface{})
	Branch(in, []chan<- interface{}{in1, in2})
	go func() {
		for i := 0; i < 10; i++ {
			in <- i
		}
		close(in)
	}()
	result1 := <-out1
	result2 := <-out2

	expected1 := (1 + 19) * 10 / 2
	expected2 := (0 + 27) * 10 / 2

	if expected1 != result1.(int) {
		t.Errorf("Expected %d but got %d", expected1, result1.(int))
	}
	if expected2 != result2.(int) {
		t.Errorf("Expected %d but got %d", expected2, result2.(int))
	}
}
