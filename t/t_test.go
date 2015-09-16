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

package t

import (
	"math"
	"testing"
	"time"
)

// Tests a day, 20 hours and 30 mins later At is parsed correctly.
func TestNextDayAndAtMatch(test *testing.T) {
	start := newTime(time.Now(), 1, 20, 30)
	day := math.Mod(float64(7+time.Now().Weekday()+2), float64(7))
	dur := nextDayAndAtMatch(start, int(day), "20:30")
	if dur != 0 {
		test.Fatalf("next run should be in 0, found %v.", dur)
	}
}

func TestNext_EachValid(test *testing.T) {
	w := &When{Each: "2h5m"}
	dur := w.Next(time.Now())
	next := time.Now().Add(2*time.Hour + 5*time.Minute).Round(time.Second)
	if !time.Now().Add(dur).Round(time.Second).Equal(next) {
		test.Fatalf("next run should happen in %v, found %v.", -time.Since(next), dur)
	}
}

func TestNext_EachInvalid(test *testing.T) {
	w := &When{Each: "2hm"}
	dur := w.Next(time.Now())
	if dur != 0 {
		test.Fatalf("next run should happen in 0, found %v.", dur)
	}
}

// Tests every 5 minutes.
func TestNext_EveryMinutes(test *testing.T) {
	w := &When{Every: Every(5).Minutes()}
	dur := w.Next(time.Now())
	next := time.Now().Add(5 * time.Minute).Round(time.Second)
	if !time.Now().Add(dur).Round(time.Second).Equal(next) {
		test.Fatalf("next run should happen in %v, found %v.", -time.Since(next), dur)
	}
}

// Tests every hour at 00:10
func TestNext_EveryHourWithAt(test *testing.T) {
	now := newTime(time.Now(), 0, 0, 40)
	w := &When{Every: Every(1).Hours(), At: "00:10"}
	dur := w.Next(now)
	next := newTime(time.Now(), 0, 1, 10)
	if !time.Now().Before(now) {
		next = next.Add(24 * time.Hour)
	}
	if !time.Now().Add(dur).Round(time.Second).Equal(next) {
		test.Fatalf("next run should happen in %v, found %v.", -time.Since(next), dur)
	}
}

// Tests every day at 21:*7
func TestNext_EveryDayWithAtMinuteWildcard(test *testing.T) {
	start := newTime(time.Now(), 0, 20, 30) // 20:30, 2 later at 01:50
	w := &When{Every: Every(1).Days(), On: Sun, At: "21:*7"}
	dur := w.Next(start)
	next := newTime(time.Now().Add(24*time.Hour), 0, 21, 37)
	if !time.Now().Before(start) {
		next = next.Add(24 * time.Hour)
	}
	if !time.Now().Add(dur).Round(time.Second).Equal(next) {
		test.Fatalf("next run should happen in %v, found %v.", -time.Since(next), dur)
	}
}

// Tests every day at **:10
func TestNext_EveryDayWithAtHourWildcard(test *testing.T) {
	start := newTime(time.Now(), 0, 0, 0) // 20:30, 2 later at 01:50
	w := &When{Every: Every(1).Days(), On: Sun, At: "**:10"}
	dur := w.Next(start)
	next := newTime(time.Now(), 0, 0, 10)
	if !time.Now().Before(start) {
		next = next.Add(24 * time.Hour)
	}
	if !time.Now().Add(dur).Round(time.Second).Equal(next) {
		test.Fatalf("next run should happen in %v, found %v.", -time.Since(next), dur)
	}
}

// Tests every day at 01:50 and on Sunday (invalid).
func TestNext_EveryDayWithAtAndDay(test *testing.T) {
	start := newTime(time.Now(), 0, 20, 30) // 20:30, 2 later at 01:50
	w := &When{Every: Every(2).Days(), On: Sun, At: "01:50"}
	dur := w.Next(start)
	untilNextSunday := time.Duration(7-time.Now().Weekday()) * time.Hour * 24
	next := newTime(time.Now().Add(24*time.Hour), 0, 1, 50)
	if untilNextSunday > 2*24*time.Hour {
		next = next.Add(2 * 24 * time.Hour)
	} else {
		next = next.Add(untilNextSunday)
	}
	if !time.Now().Before(start) {
		next = next.Add(24 * time.Hour)
	}
	if !time.Now().Add(dur).Round(time.Second).Equal(next) {
		test.Fatalf("next run should happen in %v, found %v.", -time.Since(next), dur)
	}
}

// Tests every week at 12:00 and on Sunday (invalid?).
func TestNext_EveryWithWeekAtAndDay(test *testing.T) {
	start := newTime(time.Now(), 0, 0, 0)
	w := &When{Every: Every(1).Weeks(), On: Sun, At: "12:00"}
	dur := w.Next(start)
	untilNextSunday := time.Duration(7-time.Now().Weekday()) * time.Hour * 24
	next := newTime(time.Now(), 0, 12, 00).Add(untilNextSunday)
	if !time.Now().Before(start) {
		next = next.Add(7 * 24 * time.Hour)
	}
	if !time.Now().Add(dur).Round(time.Second).Equal(next) {
		test.Fatalf("next run should happen in %v, found %v.", -time.Since(next), dur)
	}
}

func nextTime(date time.Time, daysLater, hour, min int) time.Time {
	return date.Add(time.Duration(daysLater*(24+hour))*time.Hour + time.Duration(min)*time.Minute)
}

func newTime(date time.Time, days int, exactHour, exactMin int) time.Time {
	next := date.Add(time.Duration(days) * time.Hour)
	return time.Date(next.Year(), next.Month(), next.Day(), exactHour, exactMin, 0, 0, next.Location())
}
