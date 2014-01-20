# ticktock

[![Build Status](https://travis-ci.org/rakyll/ticktock.png?branch=master)](https://travis-ci.org/rakyll/ticktock)

ticktock is a cron job scheduler that allows you to define and run periodic jobs written in Golang. ticktock also optionally provides automatic job retry if the job has failed with an error. ticktock supports delayed and repeating jobs.

**Note: Work in progress, don't use it on prod yet.**

~~~ go
// Schedule a job to email reminders once in every 3mins 10 secs.
ticktock.Schedule("email-reminders", job, &t.When{Each: "3m10s"})
ticktock.Start()
~~~

## Usage
Import and `go get` the ticktock package.

~~~ go
import "github.com/rakyll/ticktock"
~~~

The jobs you would like to schedule needs to implement `ticktock.Job` interface by providing runnable. The following example is a sample job that prints the given message.

~~~ go
type PrintJob struct {
  Msg string
}

func (j *PrintJob) Run() error {
  fmt.Println(j.Msg)
  return nil
}
~~~

### Scheduling repeated jobs

Once you've defined a Job, you need to schedule an instance of the defined job and start the scheduler. Each registered job should have a unique name, otherwise an error will be returned.

~~~ go
// Prints "Hello world" once in every seconds
err := ticktock.Schedule(
    "print-hello",
    &PrintJob{Msg: "Hello world"},
    &t.When{Every: t.Every(1).Seconds()})
~~~

If the scheduler has been started before, the job will be managed to run automatically. Otherwise, it will wait for the scheduler to be started. The scheduler can be started with the following line.

~~~ go
// typically, you schedule all jobs here and start the scheduler
ticktock.Start()
~~~

### Scheduling delayed jobs

Not all of the scheduled jobs need to run every once a while. You can also schedule a job to run at a time for only once. "Hello world" will be printed once on the next Sunday at 12:00.

~~~ go
ticktock.Schedule(
  "print-hello-once", &PrintJob{Msg: "Hello world"}, &t.When{Day: t.Sun, At: "12:00"})
~~~

### Automatic retrys

Scheduler provides automatic retry on jobs failures. In order to configure a retry count, schedule the job with additional options, providing a retry count. In the following case, we schedule the print job to be retried 2 times if it fails. (In this sample case, the job will never be retried, because `Run` always returns `nil` though.)

~~~ go
// Prints "Hello hi" once in every week, on Saturday at 10:00
ticktock.ScheduleWithOpts(
    "print-hi",
    &PrintJob{Msg: "Hello hi"},
    &t.Opts{RetryCount: 2, When: &t.When{Every: &t.Every(1).Weeks(), Day: t.Sat, At: "10:00"}})
~~~

### Cancelling jobs

Use the unique name to cancel the job. If the job is currently running, scheduler will wait for it to be completed and cancel the future runs.

~~~ go
// print-hi job will not run again
ticktock.Cancel("print-hi")
~~~

### Intervals

This section provides some valid interval samples.

~~~ go
// Every 2 minutes
t.When{Each: "2m"}

// Every 100 milliseconds
t.When{Every: t.Every(100).Milliseconds()}

// Every hour at :30
t.When{Every: t.Every(1).Hours(), At: "**:30"}

// Every day at the next beginning of an hour **:00
t.When{Every: t.Every(1).Days(), At: "**:00"}

// Every 2 weeks on Saturdays at 10:00
t.When{Every: &t.Every(2).Weeks(), On: t.Sat, At: "10:00"}

// Saturday at 15:00, not repeated
t.When{Day: t.Sat, At: "15:00"}

// Every week on Sun at 11:00, last run was explicitly given.
// If your process shuts down at 10:00 on Sunday, it allows scheduler
// to schedule the job to run in a hour on an immediate restart.
t.When{LastRun: lastRun, Every: &t.Every(1).Weeks(), On: t.Sun, At: "10:00"}
~~~

## License
Copyright 2014 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. ![Analytics](https://ga-beacon.appspot.com/UA-46881978-1/ticktock?pixel)
