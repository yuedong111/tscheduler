package main

import (
	"sync"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

)

type At struct {
	jobs       []*Job
	jobMux     sync.Mutex
	jobsWaiter sync.WaitGroup
	logger     zerolog.Logger
}

func NewAt() *At {
	at := &At{
		jobs:       make([]*Job, 0),
		logger:     log.With().Str("module", "at").Logger(),
		jobsWaiter: sync.WaitGroup{},
	}

	return at
}

func (a *At) AddJob(job *Job) {
	a.jobMux.Lock()
	defer a.jobMux.Unlock()

	a.jobs = append(a.jobs, job)
	a.jobsWaiter.Add(1)
	go a.runJob(job)
}

func (a *At) CancelJob(id uuid.UUID) {
	a.jobMux.Lock()
	defer a.jobMux.Unlock()

	a.cancelJob(id)
}

func (a *At) JobStatus(id uuid.UUID) JobResponse {
	a.jobMux.Lock()
	defer a.jobMux.Unlock()

	for _, job := range a.jobs {
		if job.ID == id {
			return JobResponse{
				At:   job.at,
				Done: job.done,
			}
		}
	}

	return JobResponse{}
}

func (a *At) runJob(job *Job) {
	now := time.Now()
	at := job.at
	job.shutdownCh = make(chan struct{})
	timer := time.NewTimer(at.Sub(now))
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			a.runAction(job)
			job.done = true

		case <-job.shutdownCh:
			a.logger.Info().
				Str("id", job.ID.String()).
				Time("at", job.at).
				Msg("run: canceled")
			job.canceled = true
		}
	}
}

func (a *At) runAction(job *Job) {
	go func() {
		job.Run(&a.jobsWaiter)
	}()
}

func (a *At) cancelJob(id uuid.UUID) {

	var jobs = make([]*Job, 0)

	for _, job := range a.jobs {
		sameID := job.ID == id
		if sameID && !job.done {
			job.canceled = true
			job.Stop()
		} else {
			jobs = append(jobs, job)
		}
	}
	a.jobs = jobs
}

func (a *At) wait() {
	a.jobsWaiter.Wait()
}
