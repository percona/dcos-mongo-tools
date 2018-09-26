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

	"github.com/percona/dcos-mongo-tools/healthcheck"
	"github.com/percona/dcos-mongo-tools/internal"
	"github.com/percona/dcos-mongo-tools/internal/db"
	"github.com/percona/dcos-mongo-tools/internal/tool"
	"github.com/percona/pmgo"
	log "github.com/sirupsen/logrus"
)

var (
	GitCommit     string
	GitBranch     string
	enableSecrets bool
)

func main() {
	app, _ := tool.New("Performs health and readiness checks for MongoDB", GitCommit, GitBranch)
	app.Flag(
		"enableSecrets",
		"enable secrets, this causes passwords to be loaded from files, overridden by env var "+internal.EnvSecretsEnabled,
	).Envar(internal.EnvSecretsEnabled).BoolVar(&enableSecrets)

	health := app.Command("health", "Run MongoDB health check")
	readiness := app.Command("readiness", "Run MongoDB readiness check").Default()
	cnf := db.NewConfig(
		app,
		internal.EnvMongoDBClusterMonitorUser,
		internal.EnvMongoDBClusterMonitorPassword,
	)

	command, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Cannot parse command line: %s", err)
	}
	if enableSecrets {
		cnf.DialInfo.Password = internal.PasswordFromFile(
			os.Getenv(internal.EnvMesosSandbox),
			cnf.DialInfo.Password,
			"password",
		)
	}

	session, err := db.GetSession(cnf)
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
			os.Exit(state.ExitCode())
		}
		log.Debugf("Member passed health check with replication state: %s", memberState)
	case readiness.FullCommand():
		log.Debug("Running readiness check")
		state, err := healthcheck.ReadinessCheck(pmgo.NewSessionManager(session))
		if err != nil {
			log.Debug(err.Error())
			session.Close()
			os.Exit(state.ExitCode())
		}
		log.Debug("Member passed readiness check")
	}
}
