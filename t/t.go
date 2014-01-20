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

// Package t contains additional scheduler options.
package t

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"
)

const (
	NoDay = iota
	Sun
	Mon
	Tue
	Wed
	Thu
	Fri
	Sat

	tNone = iota
	tMillisecond
	tSecond
	tMinute
	tHour
	tDay
	tWeek
)

// Represents options for a scheduled job.
type Opts struct {
	When *When

	RetryCount int
	Timeout    time.Duration
}

// Represents timing for schedule jobs.
// Examples:
// 		&When{Every: Every(1).Seconds()} // every seconds
// 		&When{Every: Every(2).Hours(), At: "10:00"} // every two hour at 10am
// 		&When{Every: Every(1).Hours(), At :"**:*5"} // every hour at the first *5 minute
// 		&When{Every: Every(2).Weeks(), On: Sun, At: "12:12"} // every 2 weeks on Sunday at 12:12
// 		&When{Each: "2h3m"} // every 2 hour and 3 minutes
type When struct {
	LastRun time.Time
	Each    string // string parseable by time.ParseDuration

	Every *every
	On    int
	At    string
}

type every struct {
	t int
	n int
}

// Schedules the scheduler to run repeatingly.
// If n is smaller than 1, it is set to 1.
func Every(n int) *every {
	// n should be bigger than 0
	if n < 1 {
		n = 1
	}
	return &every{t: tSecond, n: n}
}

// Sets the unit to milliseconds.
func (e *every) Milliseconds() *every {
	e.t = tMillisecond
	return e
}

// Sets the unit to Seconds.
func (e *every) Seconds() *every {
	e.t = tSecond
	return e
}

// Sets the unit to minutes.
func (e *every) Minutes() *every {
	e.t = tMinute
	return e
}

// Sets the unit to hours.
func (e *every) Hours() *every {
	e.t = tHour
	return e
}

// Sets the unit to days.
func (e *every) Days() *every {
	e.t = tDay
	return e
}

// Sets the unit to weeks.
func (e *every) Weeks() *every {
	e.t = tWeek
	return e
}

// Duration from start to the next scheduled moment.
func (w *When) Next(start time.Time) time.Duration {
	var interval, diff time.Duration
	interval = w.Duration(start)
	for {
		diff = start.Add(interval).Sub(time.Now())
		if diff > 0 {
			break
		}
		// fake the run in the past
		// and look for the next run time in the future.
		interval += w.Duration(start.Add(interval))
	}
	return diff
}

func (w *When) Duration(start time.Time) time.Duration {
	if w.Each != "" {
		dur, _ := time.ParseDuration(w.Each)
		return dur
	}
	// handle if no Every
	if w.Every == nil {
		return nextDayAndAtMatch(start, w.On, w.At)
	}

	var dur time.Duration
	n := time.Duration(w.Every.n)
	switch w.Every.t {
	case tMillisecond:
		dur = n * time.Millisecond
	case tSecond:
		dur = n * time.Second
	case tMinute:
		dur = n * time.Minute
	case tHour:
		dur = n * time.Hour
		// TODO: ignore At's hour
		dur += nextAtMatch(start, w.At)
	case tDay:
		dur = n * 24 * time.Hour
		dur += nextAtMatch(start, w.At)
	case tWeek:
		// handle Day
		weekdayDiff := 0
		if w.On != NoDay {
			weekdayDiff = int(math.Mod(float64(7+w.On-1-int(start.Weekday())), 7))
		}
		// Handle n
		dur = (n*7 + time.Duration(weekdayDiff)) * 24 * time.Hour
		// handle At
		dur += nextAtMatch(start, w.At)
	default:
		return time.Duration(0)
	}
	return dur
}

func nextDayAndAtMatch(start time.Time, day int, at string) time.Duration {
	dur := nextAtMatch(start, at)
	if day != NoDay {
		nextDay := start.Add(nextDayMatch(start, day))
		dur = nextAtMatch(nextDay, at)
	}
	return dur
}

func nextAtMatch(start time.Time, at string) (d time.Duration) {
	re := regexp.MustCompile("([\\d|\\*]{2}):([\\d|\\*]\\d)")
	matches := re.FindAllStringSubmatch(at, -1)
	if len(matches) < 1 {
		return
	}
	hour, minute := matches[0][1], matches[0][2]
	var diff time.Duration
	if hour != "**" {
		// if not, any hour, match the next minute
		diff += time.Hour * time.Duration(mod(hour, start.Hour(), 24))
	}
	if minute[0] == '*' {
		// at every minute[1] minute
		l := math.Mod(float64(start.Minute()), 10)
		d := mod(fmt.Sprintf("%c", minute[1]), int(l), 10)
		diff += time.Minute * time.Duration(d)
	} else {
		diff += time.Minute * time.Duration(mod(minute, start.Minute(), 60))
	}
	return diff
}

func nextDayMatch(start time.Time, day int) (d time.Duration) {
	diff := math.Mod(float64(7+(day-1)-int(start.Weekday())), 7)
	return time.Duration(diff) * 24 * time.Hour
}

func mod(val string, n, m int) float64 {
	value, _ := strconv.ParseInt(val, 10, 64)
	return math.Mod(float64(m+int(value)-n), float64(m))
}
