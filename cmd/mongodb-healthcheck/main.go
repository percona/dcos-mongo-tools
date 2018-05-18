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

package main

import (
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/healthcheck"
	log "github.com/sirupsen/logrus"
)

var (
	GitCommit string
	GitBranch string
)

func main() {
	app := kingpin.New("mongodb-healthcheck", "Performs DC/OS health and readiness checks for MongoDB")
	common.HandleAppVersion(app, GitCommit, GitBranch)

	health := app.Command("health", "Run DCOS health check")
	readiness := app.Command("readiness", "Run DCOS readiness check").Default()
	cnf := &healthcheck.Config{
		DB: db.NewConfig(
			app,
			common.EnvMongoDBClusterMonitorUser,
			common.EnvMongoDBClusterMonitorPassword,
		),
	}
	common.SetupLogger(app, common.GetLogFormatter(os.Args[0]), os.Stdout)

	command, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Cannot parse command line: %s", err)
	}

	session, err := db.GetSession(cnf.DB)
	if err != nil {
		log.Fatalf("Error connecting to mongodb: %s", err)
		return
	}
	defer session.Close()

	switch command {
	case health.FullCommand():
		log.Debug("Running health check")
		state, memberState, err := healthcheck.HealthCheck(session, healthcheck.OkMemberStates)
		if err != nil {
			log.Debug(err.Error())
			session.Close()
			os.Exit(int(state))
		}
		log.Debugf("Member passed health check with replication state: %s", memberState)
	case readiness.FullCommand():
		log.Debug("Running readiness check")
		state, err := healthcheck.ReadinessCheck(session)
		if err != nil {
			log.Debug(err.Error())
			session.Close()
			os.Exit(int(state))
		}
		log.Debug("Member passed readiness check")
	}
}
