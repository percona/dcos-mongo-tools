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
	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/controller"
	"github.com/percona/dcos-mongo-tools/controller/replset"
	"github.com/percona/dcos-mongo-tools/controller/user"
	log "github.com/sirupsen/logrus"
)

var (
	cmdReplset       = kingpin.Command("replset", "Control MongoDB replsets")
	cmdInit          = cmdReplset.Command("init", "Initiate a MongoDB replica set")
	cmdUser          = kingpin.Command("user", "Control MongoDB users")
	cmdUserRemove    = cmdUser.Command("remove", "Remove a MongoDB user")
	cmdUserUpdate    = cmdUser.Command("update", "Add/update a MongoDB user")
	cmdUserReloadSys = cmdUser.Command("reload-system", "Reload the DCOS Framework MongoDB system users")
)

func handleReplsetCmd(cnf *controller.Config) {
	// replset init
	cmdInit.Flag(
		"primaryAddr",
		"mongodb primary (host:port) to use to initiate the replset, overridden by env var "+common.EnvMongoDBPrimaryAddr,
	).Envar(common.EnvMongoDBPrimaryAddr).Required().StringVar(&cnf.ReplsetInit.PrimaryAddr)
	cmdInit.Flag(
		"delay",
		"amount of time to delay the init process, overridden by env var INIT_INITIATE_DELAY",
	).Default(controller.DefaultInitDelay).Envar("INIT_INITIATE_DELAY").DurationVar(&cnf.ReplsetInit.Delay)
	cmdInit.Flag(
		"maxConnectTries",
		"number of times to retry connect to mongodb, overridden by env var INIT_MAX_CONNECT_TRIES",
	).Default(controller.DefaultMaxConnectTries).Envar("INIT_MAX_CONNECT_TRIES").UintVar(&cnf.ReplsetInit.MaxConnectTries)
	cmdInit.Flag(
		"maxReplTries",
		"number of times to retry initiating mongodb replica set, overridden by env var INIT_MAX_INIT_REPLSET_TRIES",
	).Default(controller.DefaultInitMaxReplTries).Envar("INIT_MAX_INIT_REPLSET_TRIES").UintVar(&cnf.ReplsetInit.MaxReplTries)
	cmdInit.Flag(
		"maxAddUsersTries",
		"number of times to retry adding mongodb users at init time, overridden by env var INIT_MAX_ADD_USERS_TRIES",
	).Default(controller.DefaultInitMaxAddUsersTries).Envar("INIT_MAX_ADD_USERS_TRIES").UintVar(&cnf.ReplsetInit.MaxAddUsersTries)
	cmdInit.Flag(
		"retrySleep",
		"amount of time to wait between retries, overridden by env var INIT_RETRY_SLEEP",
	).Default(controller.DefaultRetrySleep).Envar("INIT_RETRY_SLEEP").DurationVar(&cnf.ReplsetInit.RetrySleep)
}

func handleUserCmd(cnf *controller.Config) {
	// user
	cmdUser.Flag(
		"endpoint",
		"DC/OS SDK service mongod endpoint name, overridden by env var "+common.EnvMongoDBMongodEndpointName,
	).Default(common.DefaultMongoDBMongodEndpointName).Envar(common.EnvMongoDBMongodEndpointName).StringVar(&cnf.User.EndpointName)
	cmdUser.Flag(
		"apiHostPrefix",
		"DC/OS SDK API hostname prefix, used to construct the DCOS API hostname",
	).Default(api.DefaultHostPrefix).StringVar(&cnf.User.API.HostPrefix)
	cmdUser.Flag(
		"apiHostSuffix",
		"DC/OS SDK API hostname suffix, used to construct the DCOS API hostname",
	).Default(api.DefaultHostSuffix).StringVar(&cnf.User.API.HostSuffix)
	cmdUser.Flag(
		"apiTimeout",
		"DC/OS SDK API timeout, overridden by env var",
	).Default(api.DefaultTimeout).DurationVar(&cnf.User.API.Timeout)
	cmdUser.Flag(
		"maxConnectTries",
		"number of times to retry connecting to mongodb",
	).Default(controller.DefaultMaxConnectTries).Envar("MONGODB_USER_CHANGE_CONNECT_TRIES").UintVar(&cnf.User.MaxConnectTries)
	cmdUser.Flag(
		"retrySleep",
		"number of times to retry connecting to mongodb",
	).Default(controller.DefaultRetrySleep).Envar("MONGODB_USER_CHANGE_RETRY_SLEEP").DurationVar(&cnf.User.RetrySleep)

	// user remove
	cmdUserRemove.Flag(
		"user",
		"the MongoDB user to be removed, system users will be skipped. this flag or env var "+common.EnvMongoDBChangeUserUsername+" is required",
	).Envar(common.EnvMongoDBChangeUserUsername).Required().StringVar(&cnf.User.Username)
	cmdUserRemove.Flag(
		"db",
		"the MongoDB database of the user, this flag or env var "+common.EnvMongoDBChangeUserDb+" is required",
	).Envar(common.EnvMongoDBChangeUserDb).Required().StringVar(&cnf.User.Database)

	// user update
	cmdUserUpdate.Arg(
		"file",
		"the required base64-encoded BSON file describing the MongoDB user to be updated",
	).Required().ExistingFileVar(&cnf.User.File)
	cmdUserUpdate.Flag(
		"db",
		"the MongoDB database of the user, this flag or env var "+common.EnvMongoDBChangeUserDb+" is required",
	).Envar(common.EnvMongoDBChangeUserDb).Required().StringVar(&cnf.User.Database)
}

func handleFailed(err error) {
	log.Fatalf("Failed with error: %s", err)
	os.Exit(1)
}

func main() {
	cnf := &controller.Config{
		Tool:        common.NewToolConfig(os.Args[0]),
		ReplsetInit: &controller.ConfigReplsetInit{},
		User: &controller.ConfigUser{
			API: &api.Config{},
		},
	}

	kingpin.Flag(
		"framework",
		"DC/OS SDK framework/service name, overridden by env var "+common.EnvFrameworkName,
	).Default(common.DefaultFrameworkName).Envar(common.EnvFrameworkName).StringVar(&cnf.FrameworkName)
	kingpin.Flag(
		"replset",
		"mongodb replica set name, this flag or env var "+common.EnvMongoDBReplset+" is required",
	).Envar(common.EnvMongoDBReplset).Required().StringVar(&cnf.Replset)
	kingpin.Flag(
		"userAdminUser",
		"mongodb userAdmin username, overridden by env var "+common.EnvMongoDBUserAdminUser,
	).Envar(common.EnvMongoDBUserAdminUser).Required().StringVar(&cnf.UserAdminUser)
	kingpin.Flag(
		"userAdminPassword",
		"mongodb userAdmin username, overridden by env var "+common.EnvMongoDBUserAdminPassword,
	).Envar(common.EnvMongoDBUserAdminPassword).Required().StringVar(&cnf.UserAdminPassword)

	handleReplsetCmd(cnf)
	handleUserCmd(cnf)

	if cnf.Tool.PrintVersion {
		cnf.Tool.PrintVersionAndExit()
	}
	common.SetupLogger(cnf.Tool)

	switch kingpin.Parse() {
	case cmdInit.FullCommand():
		err := replset.NewInitiator(cnf).Run()
		if err != nil {
			handleFailed(err)
		}
	case cmdUserUpdate.FullCommand():
		uc, err := user.NewUserController(cnf)
		if err != nil {
			handleFailed(err)
		}
		defer uc.Close()

		err = uc.UpdateUsers()
		if err != nil {
			handleFailed(err)
		}
	case cmdUserRemove.FullCommand():
		uc, err := user.NewUserController(cnf)
		if err != nil {
			handleFailed(err)
		}
		defer uc.Close()

		err = uc.RemoveUser()
		if err != nil {
			handleFailed(err)
		}
	case cmdUserReloadSys.FullCommand():
		uc, err := user.NewUserController(cnf)
		if err != nil {
			handleFailed(err)
		}
		defer uc.Close()

		err = uc.ReloadSystemUsers()
		if err != nil {
			handleFailed(err)
		}
	}
}
