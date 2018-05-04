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
	health    = kingpin.Command("health", "Run DCOS health check")
	readiness = kingpin.Command("readiness", "Run DCOS readiness check").Default()
)

func main() {
	cnf := &healthcheck.Config{
		Tool: common.NewToolConfig(os.Args[0]),
		DB: db.NewConfig(
			common.EnvMongoDBClusterMonitorUser,
			common.EnvMongoDBClusterMonitorPassword,
		),
	}
	command := kingpin.Parse()

	if cnf.Tool.PrintVersion {
		cnf.Tool.PrintVersionAndExit()
	}

	common.SetupLogger(cnf.Tool, common.GetLogFormatter(cnf.Tool.ProgName), os.Stdout)

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
