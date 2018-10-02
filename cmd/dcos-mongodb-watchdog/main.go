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
	"os/signal"
	"syscall"

	"github.com/percona/mongodb-orchestration-tools/internal/db"
	"github.com/percona/mongodb-orchestration-tools/internal/dcos"
	"github.com/percona/mongodb-orchestration-tools/internal/dcos/api"
	"github.com/percona/mongodb-orchestration-tools/internal/tool"
	"github.com/percona/mongodb-orchestration-tools/watchdog"
	config "github.com/percona/mongodb-orchestration-tools/watchdog/config"
	log "github.com/sirupsen/logrus"
)

var (
	GitCommit string
	GitBranch string
)

func main() {
	app, _ := tool.New(
		"A daemon for watching the DC/OS SDK API for MongoDB tasks and updating the MongoDB replica set state on changes",
		GitCommit, GitBranch,
	)
	cnf := &config.Config{
		DCOS: &config.ConfigDCOS{
			API: &api.Config{},
		},
	}
	app.Flag(
		"framework",
		"API framework name, this flag or env var "+dcos.EnvFrameworkName+" is required",
	).Default(dcos.DefaultFrameworkName).Envar(dcos.EnvFrameworkName).StringVar(&cnf.DCOS.FrameworkName)
	app.Flag(
		"username",
		"MongoDB clusterAdmin username, this flag or env var "+dcos.EnvMongoDBClusterAdminUser+" is required",
	).Envar(dcos.EnvMongoDBClusterAdminUser).Required().StringVar(&cnf.Username)
	app.Flag(
		"password",
		"MongoDB clusterAdmin password, this flag or env var "+dcos.EnvMongoDBClusterAdminPassword+" is required",
	).Envar(dcos.EnvMongoDBClusterAdminPassword).Required().StringVar(&cnf.Password)
	//app.Flag(
	//	"enableSecrets",
	//	"enable DC/OS Secrets, this causes passwords to be loaded from files, overridden by env var "+dcos.EnvSecretsEnabled,
	//).Envar(dcos.EnvSecretsEnabled).BoolVar(&enableSecrets)
	app.Flag(
		"apiPoll",
		"Frequency of DC/OS SDK API polls, overridden by env var WATCHDOG_API_POLL",
	).Default(config.DefaultAPIPoll).Envar("WATCHDOG_API_POLL").DurationVar(&cnf.DCOS.APIPoll)
	app.Flag(
		"apiTimeout",
		"DC/OS SDK API timeout, overridden by env var WATCHDOG_API_TIMEOUT",
	).Default(api.DefaultHTTPTimeout).Envar("WATCHDOG_API_TIMEOUT").DurationVar(&cnf.DCOS.API.Timeout)
	app.Flag(
		"ignoreAPIPods",
		"DC/OS SDK pods to ignore/exclude from watching",
	).Default(config.DefaultIgnorePods...).StringsVar(&cnf.IgnorePods)
	app.Flag(
		"replsetPoll",
		"Frequency of replset state polls or updates, overridden by env var WATCHDOG_REPLSET_POLL",
	).Default(config.DefaultReplsetPoll).Envar("WATCHDOG_REPLSET_POLL").DurationVar(&cnf.ReplsetPoll)
	app.Flag(
		"replsetTimeout",
		"MongoDB connect timeout, should be less than 'replsetPoll', overridden by env var WATCHDOG_REPLSET_TIMEOUT",
	).Default(config.DefaultReplsetTimeout).Envar("WATCHDOG_REPLSET_TIMEOUT").DurationVar(&cnf.ReplsetTimeout)
	app.Flag(
		"apiHost",
		"DC/OS SDK API hostname, overridden by env var "+dcos.EnvSchedulerAPIHost,
	).Default(api.DefaultSchedulerHost).Envar(dcos.EnvSchedulerAPIHost).StringVar(&cnf.DCOS.API.Host)
	app.Flag(
		"apiSecure",
		"Use secure connections to DC/OS SDK API",
	).BoolVar(&cnf.DCOS.API.Secure)
	app.Flag(
		"metricsPort",
		"Prometheus Metrics listen port, overridden by env var "+dcos.EnvWatchdogMetricsPort,
	).Default(watchdog.DefaultMetricsPort).Envar(dcos.EnvWatchdogMetricsPort).StringVar(&cnf.DCOS.MetricsPort)

	cnf.SSL = db.NewSSLConfig(app)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Cannot parse command line: %s", err)
	}
	//if enableSecrets {
	//	cnf.Password = internal.PasswordFromFile(
	//		os.Getenv(dcos.EnvMesosSandbox),
	//		cnf.Password,
	//		"userAdmin",
	//	)
	//}

	quit := make(chan bool)
	watchdog.New(cnf, &quit, api.New(
		cnf.DCOS.FrameworkName,
		cnf.API,
	)).Run()

	// wait for signals from the OS
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-signals
	log.Infof("Received %s signal, killing watchdog", sig)

	// send quit to all goroutines
	quit <- true
}
