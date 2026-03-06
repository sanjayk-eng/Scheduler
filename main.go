package main

import (
	"fmt"
	"sync"
	"time"
)

type Job struct {
	Name     string
	Interval time.Duration
	Task     func()
	stop     chan struct{}
}

type Scheduler struct {
	jobs []*Job
	mu   sync.Mutex
	wg   sync.WaitGroup
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		jobs: []*Job{},
	}
}

func (s *Scheduler) AddJob(name string, interval time.Duration, task func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := &Job{
		Name:     name,
		Interval: interval,
		Task:     task,
		stop:     make(chan struct{}),
	}

	s.jobs = append(s.jobs, job)
}

func (s *Scheduler) Start() {
	s.mu.Lock()
	jobsCopy := make([]*Job, len(s.jobs))
	copy(jobsCopy, s.jobs)
	s.mu.Unlock()

	for _, job := range jobsCopy {
		s.wg.Add(1)
		go s.run(job)
	}
}

func (s *Scheduler) run(job *Job) {
	defer s.wg.Done()

	for {
		start := time.Now()
		// Run the task concurrently
		go job.Task()

		// Wait for next interval or stop signal
		select {
		case <-job.stop:
			fmt.Println("Stopped:", job.Name)
			return
		case <-time.After(job.Interval - time.Since(start)):
			// continue loop
		}
	}
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	for _, job := range s.jobs {
		close(job.stop)
	}
	s.mu.Unlock()

	s.wg.Wait()
}
