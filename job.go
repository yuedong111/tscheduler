package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofrs/uuid"

)

type Action interface {
	Run()
}
type JobType string

var (
	JobOnce JobType = "once"
	JobCron JobType = "Cron"
)

type Job struct {
	ID uuid.UUID

	Action  Action
	jobType JobType

	interval   time.Duration
	at         time.Time
	canceled   bool
	done       bool
	shutdownCh chan struct{}
}

func (j *Job) Stop() {
	j.shutdownCh <- struct{}{}
}

type JobResponse struct {
	At   time.Time
	Done bool
}

type Option func(*Job)

func NewJob(opts ...Option) *Job {
	job := &Job{}
	for _, opt := range opts {
		opt(job)
	}
	return job
}

func (j *Job) Run(wg *sync.WaitGroup) {
	switch j.jobType {
	case "Cron":
		timer := time.NewTimer(j.interval)
		go func() {
			defer wg.Done()
			for {
				<-timer.C
				if j.canceled {
					timer.Stop()
					break
				}
				j.Action.Run()
				timer.Reset(j.interval)
			}
		}()
	case "once":
		j.Action.Run()
		wg.Done()
	}
}

type task struct {
	name string
	at   time.Time
}

func (t *task) Run() {
	fmt.Printf("task %s run at %s now: %s \n", t.name, t.at.Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
}
