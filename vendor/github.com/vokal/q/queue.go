package q

import (
	"runtime"
)

type Job interface {
	Run()
}

type Queue struct {
	incoming chan Job
	Quit     chan bool
}

type Worker struct {
	workers chan chan Job
	work    chan Job
}

func New(buff int) Queue {
	q := Queue{
		incoming: make(chan Job, buff),
		Quit:     make(chan bool),
	}
	return q
}

func (w *Worker) start() {
	for {
		w.workers <- w.work
		select {
		case job := <-w.work:
			job.Run()
		}
	}
}

func (q *Queue) Push(j Job) {
	q.incoming <- j
}

func (q *Queue) Start(n int) {
	workersQueue := make(chan chan Job)
	for i := 0; i < n; i++ {
		worker := Worker{
			work:    make(chan Job),
			workers: workersQueue,
		}
		go worker.start()
	}
	for {
		select {
		case work := <-q.incoming:
			go func() {
				worker := <-workersQueue
				worker <- work
			}()
		case quit := <-q.Quit:
			if quit {
				return
			}
		}
	}
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
