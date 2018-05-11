package gpool

import (
	"errors"
	"sync"
	"testing"
)

type testJob struct {
	result int
	wg     *sync.WaitGroup
}

func newTestJob(wg *sync.WaitGroup) *testJob {
	wg.Add(1)
	return &testJob{
		result: 0,
		wg:     wg,
	}
}

func (j *testJob) Do() (err error) {
	defer j.wg.Done()
	j.result = 1
	return err
}

func TestPool(t *testing.T) {
	p := NewPool(1)
	if count := p.Count(); count != 1 {
		t.Errorf("Expected 1 but got %d", count)
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
}

func TestZeroPool(t *testing.T) {
	p := NewPool(0)
	if count := p.Count(); count != 1 {
		t.Errorf("Expected 1 but got %d", count)
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
}

func TestProcJob(t *testing.T) {
	p := NewPool(1)
	var wg sync.WaitGroup
	var result int
	wg.Add(1)
	p.Channel() <- NewProcJob(func() error {
		defer wg.Done()
		result++
		return nil
	})
	wg.Add(1)
	p.Channel() <- NewProcJob(func() error {
		defer wg.Done()
		result *= 5
		return nil
	})
	wg.Wait()
	if result != 5 {
		t.Errorf("Expected 5 but got %d", result)
	}
	ex := errors.New("Expected error")
	j := NewProcJob(func() error {
		return ex
	})
	p.Channel() <- j
	if err := j.GetError(); err != ex {
		t.Errorf("Expected error %v but got %v", ex, err)
	}

	p.Shutdown()
}
