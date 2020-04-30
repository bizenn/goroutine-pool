package gpool

import (
	"sync"
	"testing"
)

type Job1 struct {
	result int
	wg     *sync.WaitGroup
}

func newTestJob(wg *sync.WaitGroup) *Job1 {
	wg.Add(1)
	return &Job1{
		result: 0,
		wg:     wg,
	}
}

func (j *Job1) Do() {
	defer j.wg.Done()
	j.result = 1
}

func (j *Job1) Result() error {
	return nil
}

func TestPoolSingleJob(t *testing.T) {
	for i, expected := range []int{1, 1, 2, 3, 4, 5, 6, 7, 8, 9} {
		p := NewPool(i)
		if count := p.Count(); count != expected {
			t.Errorf("Expected %d but got %d", expected, count)
		}
		var wg sync.WaitGroup
		j := newTestJob(&wg)
		if j.result != 0 {
			t.Errorf("Expected 0 but got %d", j.result)
		}
		p.Channel() <- j
		wg.Wait()
		if j.result != 1 {
			t.Errorf("Expected 1 but got %d", j.result)
		}
		p.Shutdown()
		p.Wait()
	}
}

type thunk struct {
	result int
	init   int
	wg     *sync.WaitGroup
	proc   func(int) int
}

func newThunk(wg *sync.WaitGroup, init int, proc func(int) int) *thunk {
	wg.Add(1)
	return &thunk{
		init: init,
		wg:   wg,
		proc: proc,
	}
}

func (t *thunk) Do() {
	t.result = t.proc(t.init)
	t.wg.Done()
}

type testEntry struct {
	expected int
	thunk    *thunk
}

func TestPoolThunk(t *testing.T) {
	p := NewPool(10)
	var wg sync.WaitGroup

	tab := []*testEntry{
		&testEntry{
			expected: 10,
			thunk:    newThunk(&wg, 1, func(init int) int { return 10 * init }),
		},
		&testEntry{
			expected: 6,
			thunk:    newThunk(&wg, 1, func(init int) int { return 5 + init }),
		},
	}

	for _, e := range tab {
		p.Channel() <- e.thunk
	}

	wg.Wait()

	for _, e := range tab {
		if e.expected != e.thunk.result {
			t.Errorf("Expected %d but got %d", e.expected, e.thunk.result)
		}
	}
	p.Shutdown()
	p.Wait()
}
