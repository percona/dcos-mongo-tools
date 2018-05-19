// Copyright 2018 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	gotesting "testing"
	"time"

	dcosmongotools "github.com/percona/dcos-mongo-tools"
	"github.com/stretchr/testify/assert"
)

func TestCommonDoStop(t *gotesting.T) {
	stop := make(chan bool)
	stopped := make(chan bool)
	go func(stop *chan bool, stopped chan bool) {
		for !DoStop(stop) {
			time.Sleep(time.Second)
		}
		stopped <- true
	}(&stop, stopped)
	stop <- true

	var tries int
	for tries < 3 {
		select {
		case _ = <-stopped:
			return
		default:
			tries += 1
		}
	}
	t.Error("Stop did not work")
}

func TestCommonDoStopFalse(t *gotesting.T) {
	stop := make(chan bool)
	stopped := make(chan bool)
	go func(stop *chan bool, stopped chan bool) {
		for !DoStop(stop) {
			time.Sleep(time.Second)
		}
		stopped <- true
	}(&stop, stopped)

	var tries int
	for tries < 3 {
		select {
		case _ = <-stopped:
			tries += 1
		default:
			stop <- true
			return
		}
	}
	t.Error("Stop did not work")
}

func TestCommonNewApp(t *gotesting.T) {
	testApp := NewApp("test help", "git-commit-here", "branch-name-here")
	appModel := testApp.Model()
	assert.Contains(t, appModel.Version, "common.test version "+dcosmongotools.Version+"\ngit commit git-commit-here, branch branch-name-here\ngo version", "kingpin.Application version is unexpected")
	assert.Equal(t, appAuthor, appModel.Author, "kingpin.Application author is unexpected")
	assert.Equal(t, "test help", appModel.Help, "kingpin.Application help is unexpected")
}
