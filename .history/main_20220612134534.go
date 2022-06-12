package main

import (
	"time"

	"github.com/gofrs/uuid"

)

func stopTaskAfter10Seconds(job *Job) {
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()
	<-timer.C
	job.Stop()
}

func main() {
	now := time.Now()
	onceTask := &task{name: "oncetask", at: time.Unix(now.Unix()+3, 0)}
	u, _ := uuid.NewV4()
	onceJob := &Job{ID: u, Action: onceTask, at: onceTask.at, jobType: "once", shutdownCh: make(chan struct{})}
	cronTask := &task{name: "crontask", at: time.Unix(now.Unix()+5, 0)}
	u1, _ := uuid.NewV4()
	cronJob := &Job{ID: u1, Action: cronTask, at: cronTask.at, jobType: "Cron", interval: 5 * time.Second, shutdownCh: make(chan struct{})}
	at := NewAt()
	at.AddJob(onceJob)
	at.AddJob(cronJob)
	go stopTaskAfter10Seconds(cronJob)
	at.wait()
}
