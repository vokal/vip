q
==========

`q` is a queue used to concurrently process jobs.
```Go
// buffer size for the queue to hold jobs on. 
Queue = q.New(buff)
```
Returns a new q.Queue.

```Go
// n is the number of workers which will
Queue.Start(n)
```
Initializes the workers, now the Queue is ready to start processing jobs.


```Go
// j is a Job, anything which implements the Run() method
Queue.Push(j)
```

Pushes jobs onto the queue.  A Job is anything thing that implements `Run()` which is responsible 
for Processing the job.


If you want to halt the Queue to prevent an additional jobs from being processed
`Queue.Quit <- true` will halt the queue.  However after the Queue has halted any Jobs that were processing will complete. 


![Q](http://upload.wikimedia.org/wikipedia/commons/thumb/6/65/Desmond_Llewelyn_01.jpg/250px-Desmond_Llewelyn_01.jpg)
