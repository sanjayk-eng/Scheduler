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

func (s *Scheduler) AddJob(
	name string,
	interval time.Duration,
	task func(),
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := &Job{
		Name:     name,
		Interval: interval,
		Task:     task,
		stop:     make(chan struct{}),
	}

	s.jobs = append(s.jobs, job)

	fmt.Println("Job added:", name)
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

	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()

	fmt.Println("Job started:", job.Name)

	// Run immediately when the job starts.
	go job.Task()

	for {
		select {

		case <-job.stop:
			fmt.Println("Job stopped:", job.Name)
			return

		case <-ticker.C:
			// Run task concurrently.
			go job.Task()
		}
	}
}

func (s *Scheduler) Stop() {
	s.mu.Lock()

	for _, job := range s.jobs {
		select {
		case <-job.stop:
			// Already stopped.
		default:
			close(job.stop)
		}
	}

	s.mu.Unlock()

	s.wg.Wait()

	fmt.Println("Scheduler stopped")
	fmt.Println("=========")
}

func main() {

	scheduler := NewScheduler()

	// Job 1
	scheduler.AddJob(
		"Email Job",
		2*time.Second,
		func() {
			fmt.Println(
				"Sending emails:",
				time.Now().Format("15:04:05"),
			)

			time.Sleep(500 * time.Millisecond)

			fmt.Println("Email job completed")
		},
	)

	// Job 2
	scheduler.AddJob(
		"Database Backup",
		5*time.Second,
		func() {
			fmt.Println(
				"Running database backup:",
				time.Now().Format("15:04:05"),
			)

			time.Sleep(1 * time.Second)

			fmt.Println("Database backup completed")
		},
	)

	// Job 3
	scheduler.AddJob(
		"Health Check",
		3*time.Second,
		func() {
			fmt.Println(
				"Health check:",
				time.Now().Format("15:04:05"),
			)
		},
	)

	// Start all jobs
	scheduler.Start()

	// Let scheduler run for 15 seconds
	time.Sleep(15 * time.Second)
	fmt.Println("run...")

	// Stop all jobs
	scheduler.Stop()

	fmt.Println("Application finished")
}