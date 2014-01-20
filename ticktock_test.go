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

package ticktock

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/rakyll/ticktock/t"
)

type counterJob struct {
	Count int
}

func (job *counterJob) Run() error {
	job.Count += 1
	return nil
}

type anyJob struct {
	Fn func()
}

func (any *anyJob) Run() error {
	any.Fn()
	return nil
}

type errorJob struct {
	count      int
	errorAfter int
}

func (job *errorJob) Run() error {
	job.count += 1
	if job.count < job.errorAfter {
		return errors.New("fake error")
	}
	return nil
}

// Tests registering two jobs with the same name.
func TestSchedule_Duplicate(test *testing.T) {
	sh := &Scheduler{}
	if err := sh.Schedule("print", &counterJob{}, &t.When{At: "**:15"}); err != nil {
		test.Fatalf("error during duplication test: %v", err)
	}
	if err := sh.Schedule("print", &counterJob{}, &t.When{At: "**:15"}); err == nil {
		test.Fatal("error expected during duplication, but not found")
	}
}

// Tests if repeating job is running on time.
func TestSchedule_OntimeRepeating(test *testing.T) {
	sh := &Scheduler{}
	var wg sync.WaitGroup

	wg.Add(1)
	job := &counterJob{}
	sh.Schedule("hi", job, &t.When{Every: t.Every(100).Milliseconds()})
	time.AfterFunc(300*time.Millisecond, func() {
		defer wg.Done()
		sh.Cancel("hi")
		if job.Count < 2 {
			test.Fatalf("scheduler worked for %v, expected to run 3 times", job.Count)
		}
	})
	go sh.Start()
	wg.Wait()
}

// Tests if jobs are starting before starting the scheduler.
func TestStart_NoStart(test *testing.T) {
	sh := &Scheduler{}
	var wg sync.WaitGroup

	wg.Add(1)
	job1 := &counterJob{}
	job2 := &counterJob{}
	sh.Schedule("hello", job1, &t.When{Every: t.Every(200).Milliseconds()})
	sh.Schedule("hi", job2, &t.When{Every: t.Every(100).Milliseconds()})
	time.AfterFunc(300*time.Millisecond, func() {
		defer wg.Done()
		if job1.Count+job2.Count > 1 {
			test.Fatalf("scheduler not started but jobs have run")
		}
	})
	wg.Wait()
}

// Tests if a job is being scheduled if it's registered after the start.
func TestStart_AfterStart(test *testing.T) {
	sh := &Scheduler{}
	var wg sync.WaitGroup

	wg.Add(1)
	job := &counterJob{}
	sh.Start()
	sh.Schedule("hi", job, &t.When{Every: t.Every(100).Milliseconds()})
	time.AfterFunc(300*time.Millisecond, func() {
		defer wg.Done()
		sh.Cancel("hi")
		if job.Count == 0 {
			test.Fatalf("job is expected to run even though it's scheduled after Start, but it didn't")
		}
	})
	wg.Wait()
}

// Tests if jobs are being cancelled.
func TestCancel(test *testing.T) {
	sh := &Scheduler{}
	var wg sync.WaitGroup

	wg.Add(1)
	job := &counterJob{}
	sh.Schedule("hi", job, &t.When{Every: t.Every(100).Milliseconds()})
	time.AfterFunc(100*time.Millisecond, func() {
		defer wg.Done()
		sh.Cancel("hi")
		if job.Count > 2 {
			test.Fatalf("scheduler cancelled but job worked more than expected")
		}
	})
	sh.Start()
	wg.Wait()
}

// Tests if job is retried if it fails.
func TestRetryCount(test *testing.T) {
	sh := &Scheduler{}
	var wg sync.WaitGroup

	wg.Add(1)
	job := &errorJob{errorAfter: 2}
	sh.ScheduleWithOpts("hi", job, &t.Opts{
		RetryCount: 2,
		When:       &t.When{Every: t.Every(100).Milliseconds()},
	})
	time.AfterFunc(200*time.Millisecond, func() {
		defer wg.Done()
		sh.Cancel("hi")
		if job.count < 2 {
			test.Fatalf("expected to retry for 2 times, ram %v times", job.count)
		}
	})
	sh.Start()
	wg.Wait()
}

// Tests if when is successfully validated when it's nil.
func TestWhen_None(test *testing.T) {
	sh := &Scheduler{}
	if err := sh.Schedule("hi", nil, nil); err == nil {
		test.Fatalf("opts.When is nil, but no error returned")
	}
}

// Tests if when is successfully validated when it's invalid.
func TestWhen_NotValid(test *testing.T) {
	sh := &Scheduler{}
	if err := sh.Schedule("hi", nil, &t.When{}); err == nil {
		test.Fatalf("opts.When is not valid, but no error returned")
	}
}

func TestWhen_LastRun(test *testing.T) {
	start := time.Now()
	sh := &Scheduler{}
	lastRun := time.Now().Add(-1000 * time.Millisecond)
	sh.Schedule("jobs-with-lastrun",
		&anyJob{Fn: func() {
			// should run in 100ms
			diff := time.Now().Sub(start)
			if diff > 205*time.Millisecond {
				test.Errorf("Job expected to run in 200ms, but it happened in %v", diff)
			}
		}},
		&t.When{LastRun: lastRun, Each: "300ms"})
	sh.Start()
}
