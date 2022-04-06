package controllers

import (
	"github.com/ringtail/go-cron"
	log "k8s.io/klog/v2"
	"time"
)

const (
	maxOutOfDateTimeout = time.Minute * 5
)

type ScalerExecutor interface {
	Run()
	Stop()
	AddJob(job ScalerJob) error
	Update(job ScalerJob) error
	RemoveJob(job ScalerJob) error
	FindJob(job ScalerJob) bool
	ListEntries() []*cron.Entry
}

type ScalerExecutorHPA struct {
	Engine *cron.Cron
}

func (se *ScalerExecutorHPA) AddJob(job ScalerJob) error {
	err := se.Engine.AddJob(job.SchedulePlan(), job)
	if err != nil {
		log.Errorf("Failed to add job to engine,because of %v", err)
	}
	return err
}

func (se *ScalerExecutorHPA) ListEntries() []*cron.Entry {
	entries := se.Engine.Entries()
	return entries
}

func (se *ScalerExecutorHPA) FindJob(job ScalerJob) bool {
	entries := se.Engine.Entries()
	for _, e := range entries {
		if e.Job.ID() == job.ID() {
			// clean up out of date jobs when it reach maxOutOfDateTimeout
			if e.Next.Add(maxOutOfDateTimeout).After(time.Now()) {
				return true
			}
			log.Warningf("The job %s is out of date and need to be clean up.", job.Name())
		}
	}
	return false
}

func (se *ScalerExecutorHPA) Update(job ScalerJob) error {
	se.Engine.RemoveJob(job.ID())
	err := se.Engine.AddJob(job.SchedulePlan(), job)
	if err != nil {
		log.Errorf("Failed to update job to engine,because of %v", err)
	}
	return err
}

func (se *ScalerExecutorHPA) RemoveJob(job ScalerJob) error {
	se.Engine.RemoveJob(job.ID())
	return nil
}

func (se *ScalerExecutorHPA) Run() {
	se.Engine.Start()
}

func (se *ScalerExecutorHPA) Stop() {
	se.Engine.Stop()
}

func NewScalerHPAExecutor(timezone *time.Location, handler func(job *cron.JobResult)) ScalerExecutor {
	if timezone == nil {
		timezone = time.Now().Location()
	}
	c := &ScalerExecutorHPA{
		Engine: cron.NewWithLocation(timezone),
	}
	c.Engine.AddResultHandler(handler)
	return c
}
