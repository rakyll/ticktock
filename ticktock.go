// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package ticktock contains a timer based job scheduler with
// automatic retry on failures.
package ticktock

import (
	"errors"
	"sync"
	"time"

	"github.com/rakyll/ticktock/t"
)

var (
	defaultScheduler = &Scheduler{}
)

// Job implements a schedulable job that implements a runnable.
type Job interface {
	Run() error
}

// Scheduler represents a job scheduler that manages
// a set of scheduled jobs.
type Scheduler struct {
	jobs    map[string]*jobC
	started bool

	wg sync.WaitGroup
	mu sync.Mutex
}

// Schedules a job called name, with the provided timing
// information. name should be unique for each scheduled job.
func Schedule(name string, job Job, when *t.When) error {
	return defaultScheduler.Schedule(name, job, when)
}

// Schedules a job called anme, with the provided options. Name
// should be unique among all scheduled jobs.
func ScheduleWithOpts(name string, job Job, opts *t.Opts) (err error) {
	return defaultScheduler.ScheduleWithOpts(name, job, opts)
}

// Cancels a scheduled job registered on the default scheduler.
// If job is already running, waits for the run to be completed
// and cancels the next runs.
func Cancel(name string) {
	defaultScheduler.Cancel(name)
}

// Starts the jobs registered for the default scheduler.
func Start() {
	defaultScheduler.Start()
}

// Schedules a job on the scheduler. Name should be unique
// among all registered jobs.
func (s *Scheduler) Schedule(name string, job Job, when *t.When) error {
	return s.ScheduleWithOpts(name, job, &t.Opts{When: when})
}

func (s *Scheduler) ScheduleWithOpts(name string, job Job, opts *t.Opts) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.jobs[name]; ok {
		return errors.New("a job already exists with the name provided")
	}
	if opts.When == nil || opts.When.Duration(time.Now()) == 0 {
		return errors.New("not a valid opts.When is provided")
	}
	if s.jobs == nil {
		s.jobs = make(map[string]*jobC)
	}
	s.jobs[name] = &jobC{
		scheduler:  s,
		job:        job,
		retryCount: opts.RetryCount,
		when:       opts.When,
		forever:    opts.When.Every != nil,
		cancelSig:  make(chan bool),
	}
	if s.started {
		s.wg.Add(1)
		s.jobs[name].schedule()
	}
	return
}

// Cancels a job called name. If there is no such job, returns
// immediately. If the job is alreading running, scheduler waits
// for the job to complete and cancels the job.
func (s *Scheduler) Cancel(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job, ok := s.jobs[name]
	if !ok {
		return
	}
	job.cancel()
	delete(s.jobs, name)
}

// Starts to schedule the jobs.
func (s *Scheduler) Start() {
	s.started = true
	for _, j := range s.jobs {
		s.wg.Add(1)
		j.schedule()
	}
	s.wg.Wait()
}

type jobC struct {
	scheduler  *Scheduler
	job        Job
	retryCount int
	when       *t.When
	forever    bool
	timer      *time.Timer
	cancelSig  chan bool
}

func (j *jobC) schedule() {
	select {
	case <-j.cancelSig:
		// TODO: cancel the timer
		j.timer.Stop()
		j.done()
		return
	default:
		if j.when.LastRun.IsZero() {
			j.when.LastRun = time.Now()
		}
		dur := j.when.Next(j.when.LastRun)
		j.timer = time.AfterFunc(dur, func() {
			j.run()
			j.when.LastRun = time.Now()
			if j.forever {
				j.schedule()
				return
			}
			j.done()
		})
	}
}

func (j *jobC) run() {
retryLoop:
	for i := 0; i < j.retryCount+1; i++ {
		if err := j.job.Run(); err == nil {
			break retryLoop
		}
	}
}

func (j *jobC) cancel() {
	j.cancelSig <- true
	if j.timer != nil {
		j.timer.Stop()
	}
}

func (j *jobC) done() {
	// TODO: jobC should not have its scheduler
	// handle this elsewhere
	j.scheduler.wg.Done()
}
