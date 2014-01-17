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

package jobs

import (
	"os/exec"
)

// CmdJob spawns a process.
// Example usage:
// cmd := exec.Command("echo", "Hello world")
// ticktock.Schedule(
//     "echo",
//     &jobs.CmdJob{Cmd: cmd},
//     &t.When{Every: t.Every(1).Seconds()})
type CmdJob struct {
	Cmd *exec.Cmd
}

// Runs the command.
func (j *CmdJob) Run() error {
	return j.Cmd.Run()
}
