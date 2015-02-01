package workerqueue

type Job interface {
	Run()
}

type Queue struct {
	WorkQueue chan Job
	Quit      chan bool
}

type Worker struct {
	WorkerQueue chan chan Job
	Work        chan Job
}

func MakeQueue(buff int) Queue {
	q := Queue{
		WorkQueue: make(chan Job, buff),
		Quit:      make(chan bool),
	}
	return q
}

func (w *Worker) start() {
	go func() {
		for {
			w.WorkerQueue <- w.Work
			select {
			case job := <-w.Work:
				job.Run()
			}
		}
	}()
}

func (q *Queue) AddJob(j Job) {
	q.WorkQueue <- j
}

func (q *Queue) Start(n int) {
	wq := make(chan chan Job)
	for i := 0; i < n; i++ {
		worker := Worker{
			Work:        make(chan Job),
			WorkerQueue: wq,
		}
		worker.start()
	}
	for {
		select {
		case work := <-q.WorkQueue:
			go func() {
				worker := <-wq
				worker <- work
			}()
		case quit := <-q.Quit:
			if quit {
				return
			}
		}
	}
}
