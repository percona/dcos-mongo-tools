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

	"github.com/alecthomas/kingpin"
	"github.com/percona/dcos-mongo-tools/executor"
	"github.com/percona/dcos-mongo-tools/executor/config"
	"github.com/percona/dcos-mongo-tools/executor/job"
	"github.com/percona/dcos-mongo-tools/executor/metrics"
	"github.com/percona/dcos-mongo-tools/executor/mongodb"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
	"github.com/percona/dcos-mongo-tools/internal"
	"github.com/percona/dcos-mongo-tools/internal/db"
	"github.com/percona/dcos-mongo-tools/internal/tool"
	log "github.com/sirupsen/logrus"
)

var (
	GitCommit     string
	GitBranch     string
	enableSecrets bool
)

func handleMongoDB(app *kingpin.Application, cnf *config.Config) {
	app.Flag(
		"mongodb.totalMemoryMB",
		"the total amount of system memory, in megabytes",
	).Envar(internal.EnvMongoDBMemoryMB).Required().UintVar(&cnf.MongoDB.TotalMemoryMB)
	app.Flag(
		"mongodb.configDir",
		"path to mongodb instance config file, defaults to $"+internal.EnvMesosSandbox+" if available, otherwise "+mongodb.DefaultConfigDirFallback,
	).Default(mongodb.DefaultConfigDirFallback).Envar(internal.EnvMesosSandbox).StringVar(&cnf.MongoDB.ConfigDir)
	app.Flag(
		"mongodb.binDir",
		"path to mongodb binary directory",
	).Default(mongodb.DefaultBinDir).StringVar(&cnf.MongoDB.BinDir)
	app.Flag(
		"mongodb.tmpDir",
		"path to mongodb temporary directory, defaults to $"+internal.EnvMesosSandbox+"/tmp if available, otherwise "+mongodb.DefaultTmpDirFallback,
	).Default(config.MesosSandboxPathOrFallback(
		"tmp",
		mongodb.DefaultTmpDirFallback,
	)).StringVar(&cnf.MongoDB.TmpDir)
	app.Flag(
		"mongodb.user",
		"user to run mongodb instance as",
	).Default(mongodb.DefaultUser).StringVar(&cnf.MongoDB.User)
	app.Flag(
		"mongodb.group",
		"group to run mongodb instance as",
	).Default(mongodb.DefaultGroup).StringVar(&cnf.MongoDB.Group)
	app.Flag(
		"mongodb.wiredTigerCacheRatio",
		"the ratio of system memory to be used for wiredTiger cache",
	).Default(mongodb.DefaultWiredTigerCacheRatio).Envar(internal.EnvMongoDBWiredTigerCacheSizeRatio).Float64Var(&cnf.MongoDB.WiredTigerCacheRatio)
}

func handleMetrics(app *kingpin.Application, cnf *config.Config) {
	app.Flag(
		"metrics.enable",
		"Enable DC/OS Metrics monitoring for MongoDB, defaults to "+internal.EnvMetricsEnabled+" env var",
	).Envar(internal.EnvMetricsEnabled).BoolVar(&cnf.Metrics.Enabled)
	app.Flag(
		"metrics.interval",
		"The frequency to send metrics to DC/OS Metrics service, defaults to "+internal.EnvMetricsInterval+" env var",
	).Default(metrics.DefaultInterval).Envar(internal.EnvMetricsInterval).DurationVar(&cnf.Metrics.Interval)
	app.Flag(
		"metrics.statsd_host",
		"The frequency to send metrics to DC/OS Metrics service, defaults to "+internal.EnvMetricsStatsdHost+" env var",
	).Envar(internal.EnvMetricsStatsdHost).StringVar(&cnf.Metrics.StatsdHost)
	app.Flag(
		"metrics.statsd_port",
		"The frequency to send metrics to DC/OS Metrics service, defaults to "+internal.EnvMetricsStatsdPort+" env var",
	).Envar(internal.EnvMetricsStatsdPort).IntVar(&cnf.Metrics.StatsdPort)
}

func handlePmm(app *kingpin.Application, cnf *config.Config) {
	app.Flag(
		"pmm.configDir",
		"Directory containing the PMM client config file (pmm.yml), defaults to "+internal.EnvMesosSandbox+" env var",
	).Envar(internal.EnvMesosSandbox).StringVar(&cnf.PMM.ConfigDir)
	app.Flag(
		"pmm.enable",
		"Enable Percona PMM monitoring for OS and MongoDB, defaults to "+internal.EnvPMMEnabled+" env var",
	).Envar(internal.EnvPMMEnabled).BoolVar(&cnf.PMM.Enabled)
	app.Flag(
		"pmm.enableQueryAnalytics",
		"Enable Percona PMM query analytics (QAN) client/agent, defaults to "+internal.EnvPMMEnableQueryAnalytics+" env var",
	).Envar(internal.EnvPMMEnableQueryAnalytics).BoolVar(&cnf.PMM.EnableQueryAnalytics)
	app.Flag(
		"pmm.serverAddress",
		"Percona PMM server address, defaults to "+internal.EnvPMMServerAddress+" env var",
	).Envar(internal.EnvPMMServerAddress).StringVar(&cnf.PMM.ServerAddress)
	app.Flag(
		"pmm.clientName",
		"Percona PMM client address, defaults to "+internal.EnvTaskName+" env var",
	).Envar(internal.EnvTaskName).StringVar(&cnf.PMM.ClientName)
	app.Flag(
		"pmm.serverSSL",
		"Enable SSL communication between Percona PMM client and server, defaults to "+internal.EnvPMMServerSSL+" env var",
	).Envar(internal.EnvPMMServerSSL).BoolVar(&cnf.PMM.ServerSSL)
	app.Flag(
		"pmm.serverInsecureSSL",
		"Enable insecure SSL communication between Percona PMM client and server, defaults to "+internal.EnvPMMServerInsecureSSL+" env var",
	).Envar(internal.EnvPMMServerInsecureSSL).BoolVar(&cnf.PMM.ServerInsecureSSL)
	app.Flag(
		"pmm.linuxMetricsExporterPort",
		"Port number for bind Percona PMM Linux Metrics exporter to, defaults to "+internal.EnvPMMLinuxMetricsExporterPort+" env var",
	).Envar(internal.EnvPMMLinuxMetricsExporterPort).UintVar(&cnf.PMM.LinuxMetricsExporterPort)
	app.Flag(
		"pmm.mongodbMetricsExporterPort",
		"Port number for bind Percona PMM MongoDB Metrics exporter to, defaults to "+internal.EnvPMMMongoDBMetricsExporterPort+" env var",
	).Envar(internal.EnvPMMMongoDBMetricsExporterPort).UintVar(&cnf.PMM.MongoDBMetricsExporterPort)
	app.Flag(
		"pmm.mongodb.clusterName",
		"Percona PMM client mongodb cluster name, defaults to "+internal.EnvFrameworkName+" env var",
	).Envar(internal.EnvFrameworkName).StringVar(&cnf.PMM.MongoDB.ClusterName)
}

func main() {
	app, verbose := tool.New("Handles running MongoDB instances and various in-container background tasks", GitCommit, GitBranch)
	app.Command("mongod", "run a mongod instance")
	app.Command("mongos", "run a mongos instance")

	dbConfig := db.NewConfig(
		app,
		internal.EnvMongoDBClusterMonitorUser,
		internal.EnvMongoDBClusterMonitorPassword,
	)
	cnf := &config.Config{
		DB:      dbConfig,
		MongoDB: &mongodb.Config{},
		Metrics: &metrics.Config{
			DB: dbConfig,
		},
		PMM: &pmm.Config{
			DB:      dbConfig,
			MongoDB: &pmm.ConfigMongoDB{},
		},
		Verbose: verbose,
	}

	app.Flag(
		"framework",
		"dcos framework name, overridden by env var "+internal.EnvFrameworkName,
	).Default(internal.DefaultFrameworkName).Envar(internal.EnvFrameworkName).StringVar(&cnf.FrameworkName)
	app.Flag(
		"connectRetrySleep",
		"duration to wait between retries of the connection/ping to mongodb",
	).Default(config.DefaultConnectRetrySleep).DurationVar(&cnf.ConnectRetrySleep)
	app.Flag(
		"delayBackgroundJobs",
		"Amount of time to delay running of executor background jobs",
	).Default(config.DefaultDelayBackgroundJob).DurationVar(&cnf.DelayBackgroundJob)
	app.Flag(
		"enableSecrets",
		"enable secrets, this causes passwords to be loaded from files, overridden by env var "+internal.EnvSecretsEnabled,
	).Envar(internal.EnvSecretsEnabled).BoolVar(&enableSecrets)

	handleMongoDB(app, cnf)
	handleMetrics(app, cnf)
	handlePmm(app, cnf)

	nodeType, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Cannot parse command line: %s", err)
	}
	cnf.NodeType = config.NodeType(nodeType)
	if enableSecrets {
		cnf.DB.DialInfo.Password = internal.PasswordFromFile(
			os.Getenv(internal.EnvMesosSandbox),
			cnf.DB.DialInfo.Password,
			"password",
		)
	}

	quit := make(chan bool, 1)
	e := executor.New(cnf, &quit)

	var daemon executor.Daemon
	daemonState := make(chan *os.ProcessState, 1)

	switch cnf.NodeType {
	case config.NodeTypeMongod:
		daemon = mongodb.NewMongod(cnf.MongoDB, daemonState)
	case config.NodeTypeMongos:
		log.Fatalf("mongos nodes are not supported yet!")
	default:
		log.Fatalf("did not start anything, this is unexpected")
	}

	// start the daemon
	err = e.Run(daemon)
	if err != nil {
		log.Fatalf("Failed to start %s daemon: %s", daemon.Name(), err)
	}

	// wait for Daemon to become available
	session, err := db.WaitForSession(
		cnf.DB,
		0,
		cnf.ConnectRetrySleep,
	)
	if err != nil {
		log.Fatalf("Error creating db session: %s", err.Error())
	}
	defer session.Close()

	// start job Runner
	go job.New(cnf, session, &quit).Run()

	// wait for signals from the OS
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// wait for OS signal or daemonState (*os.ProcessState from daemon process)
	select {
	case state := <-daemonState:
		quit <- true

		logFields := log.Fields{
			"success": state.Success(),
			"exited":  state.Exited(),
		}

		if state.String() == "exit status 0" {
			log.WithFields(logFields).Infof("%s cleanly exited with status: %s", daemon.Name(), state.String())
			os.Exit(0)
		}

		log.WithFields(logFields).Fatalf("Unexpected die/exit from %s with status: %s", daemon.Name(), state.String())
	case sig := <-signals:
		quit <- true
		log.Infof("Received %s signal, killing %s daemon and jobs", sig, daemon.Name())
	}
}
