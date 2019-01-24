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

package watcher

import (
	"os"
	"testing"
	"time"

	"github.com/percona/mongodb-orchestration-tools/internal/db"
	"github.com/percona/mongodb-orchestration-tools/internal/logger"
	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	"github.com/percona/mongodb-orchestration-tools/watchdog/config"
	"github.com/percona/mongodb-orchestration-tools/watchdog/replset"
)

var (
	testManager *WatcherManager
	testConfig  = &config.Config{
		Username:    testutils.MongodbAdminUser,
		Password:    testutils.MongodbAdminPassword,
		ReplsetPoll: 500 * time.Millisecond,
		SSL:         &db.SSLConfig{},
	}
	testWatchRs = replset.New(testConfig, testutils.MongodbReplsetName)
)

func TestMain(m *testing.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter(), os.Stdout)
	os.Exit(m.Run())
}
