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

package watchdog

import (
	"os"
	"testing"
	"time"

	"github.com/percona/mongodb-orchestration-tools/internal/db"
	"github.com/percona/mongodb-orchestration-tools/internal/dcos/api/mocks"
	"github.com/percona/mongodb-orchestration-tools/internal/logger"
	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod/dcos"
	"github.com/percona/mongodb-orchestration-tools/watchdog/config"
	"github.com/stretchr/testify/assert"
)

var (
	testWatchdog *Watchdog
	testQuitChan = make(chan bool)
	testConfig   = &config.Config{
		ServiceName: "test",
		Username:    testutils.MongodbAdminUser,
		Password:    testutils.MongodbAdminPassword,
		APIPoll:     time.Millisecond * 100,
		ReplsetPoll: time.Millisecond * 100,
		SSL:         &db.SSLConfig{},
	}
	testAPIClient = &mocks.Client{}
)

func TestMain(m *testing.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter("test"), os.Stdout)
	os.Exit(m.Run())
}

func TestWatchdogNew(t *testing.T) {
	testWatchdog = New(testConfig, &testQuitChan, testAPIClient)
	assert.NotNil(t, testWatchdog, ".New() returned nil")
}

func TestWatchdogDoIgnorePod(t *testing.T) {
	testConfig.IgnorePods = []string{"ignore-me"}
	assert.True(t, testWatchdog.doIgnorePod("ignore-me"))
	assert.False(t, testWatchdog.doIgnorePod("dont-ignore-me"))
}

func TestWatchdogRun(t *testing.T) {
	testAPIClient.On("Name").Return("test")
	testAPIClient.On("URL").Return("http://test")
	testAPIClient.On("Pods").Return([]string{"test"}, nil)

	tasks := make([]pod.Task, 0)
	tasks = append(tasks, dcos.NewTask(&dcos.TaskData{
		Info: &dcos.TaskInfo{},
	}, "test"))
	testAPIClient.On("GetTasks", "test").Return(tasks, nil)

	go testWatchdog.Run()
	time.Sleep(time.Millisecond * 150)
	close(testQuitChan)
}
