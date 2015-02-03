package q

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

var (
	test []bool
	wg   sync.WaitGroup
)

type TestJob struct {
	index int
}

func (t *TestJob) Run() {
	time.Sleep(time.Duration(rand.Intn(1)) * time.Second)
	test[t.index] = true
	wg.Done()
}

func TestQueue(t *testing.T) {
	test = []bool{false, false, false, false, false, false, false, false}
	q := New(14)
	go q.Start(2)
	for i := 0; i < len(test); i++ {
		wg.Add(1)
		j := &TestJob{index: i}
		q.Push(j)
	}
	wg.Wait()
	for i := 0; i < len(test); i++ {
		if !test[i] {
			t.Error("Test Failed: Worker failed to trigger job")
		}
	}
	q.Quit <- true
}
