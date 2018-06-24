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

package replset

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/controller"
	"github.com/percona/dcos-mongo-tools/controller/user"
	log "github.com/sirupsen/logrus"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
	"gopkg.in/mgo.v2"
)

const (
	ErrMsgDNSNotReady = "No host described in new configuration 1 for replica set rs maps to this node"
)

type Initiator struct {
	config        *controller.Config
	hostname      string
	connectTries  uint
	replInitTries uint
	usersTries    uint
}

func NewInitiator(config *controller.Config) *Initiator {
	return &Initiator{
		config: config,
	}
}

func (i *Initiator) initReplset(session *mgo.Session) error {
	rsCnfMan := rsConfig.New(session)
	if rsCnfMan.IsInitiated() {
		return errors.New("Replset should not be initiated already! Exiting")
	}

	config := rsConfig.NewConfig(i.config.Replset)
	member := rsConfig.NewMember(i.config.ReplsetInit.PrimaryAddr)
	member.Tags = &rsConfig.ReplsetTags{
		"dcosFramework": i.config.FrameworkName,
	}
	config.AddMember(member)
	rsCnfMan.Set(config)

	log.Info("Initiating replset")
	fmt.Println(config)

	for i.replInitTries <= i.config.ReplsetInit.MaxReplTries {
		err := rsCnfMan.Initiate()
		if err == nil {
			log.WithFields(log.Fields{
				"version": config.Version,
			}).Info("Initiated replset with config:")
			break
		}
		if err.Error() != ErrMsgDNSNotReady {
			log.WithFields(log.Fields{
				"replset": i.config.Replset,
				"error":   err,
			}).Error("Error initiating replset! Retrying")
		}
		time.Sleep(i.config.ReplsetInit.RetrySleep)
		i.replInitTries += 1
	}
	if i.replInitTries >= i.config.ReplsetInit.MaxReplTries {
		return errors.New("Could not init replset")
	}

	return nil
}

func (i *Initiator) initAdminUser(session *mgo.Session) error {
	err := user.UpdateUser(session, user.UserAdmin, "admin")
	if err != nil {
		log.Errorf("Error adding admin user: %s", err)
		return err
	}
	return nil
}

func (i *Initiator) initUsers(session *mgo.Session) error {
	err := user.UpdateUsers(session, user.SystemUsers, "admin")
	if err != nil {
		log.Errorf("Error adding system users: %s", err)
		return err
	}
	return nil
}

func (i *Initiator) Run() error {
	log.WithFields(log.Fields{
		"framework": i.config.FrameworkName,
	}).Info("Mongod replset initiator started")

	log.WithFields(log.Fields{
		"sleep": i.config.ReplsetInit.Delay,
	}).Info("Waiting to start initiation")
	time.Sleep(i.config.ReplsetInit.Delay)

	// use an insecure SSL connection to avoid hostname validation error for the server hostname
	sslCnfInsecure := *i.config.SSL
	sslCnfInsecure.Insecure = true

	split := strings.SplitN(i.config.ReplsetInit.PrimaryAddr, ":", 2)
	localhostNoAuthSession, err := db.WaitForSession(
		&db.Config{
			DialInfo: &mgo.DialInfo{
				Addrs:    []string{"localhost:" + split[1]},
				Direct:   true,
				FailFast: true,
				Timeout:  db.DefaultMongoDBTimeoutDuration,
			},
			SSL: &sslCnfInsecure,
		},
		i.config.ReplsetInit.MaxConnectTries,
		i.config.ReplsetInit.RetrySleep,
	)
	if err != nil {
		return err
	}
	defer localhostNoAuthSession.Close()

	log.WithFields(log.Fields{
		"host":    "localhost",
		"auth":    false,
		"replset": "",
	}).Info("Connected to MongoDB")

	err = i.initReplset(localhostNoAuthSession)
	if err != nil {
		return err
	}

	log.Info("Waiting for host to become primary")
	err = db.WaitForPrimary(localhostNoAuthSession, i.config.ReplsetInit.MaxConnectTries, i.config.ReplsetInit.RetrySleep)
	if err != nil {
		return err
	}

	err = i.initAdminUser(localhostNoAuthSession)
	if err != nil {
		return err
	}

	log.Info("Closing localhost connection, reconnecting with a replset+auth connection")
	localhostNoAuthSession.Close()

	replsetAuthSession, err := db.WaitForSession(
		&db.Config{
			DialInfo: &mgo.DialInfo{
				Addrs:          []string{i.config.ReplsetInit.PrimaryAddr},
				Username:       i.config.UserAdminUser,
				Password:       i.config.UserAdminPassword,
				ReplicaSetName: i.config.Replset,
				Direct:         false,
				FailFast:       true,
				Timeout:        db.DefaultMongoDBTimeoutDuration,
			},
			SSL: i.config.SSL,
		},
		i.config.ReplsetInit.MaxConnectTries,
		i.config.ReplsetInit.RetrySleep,
	)
	if err != nil {
		return err
	}
	defer replsetAuthSession.Close()
	log.WithFields(log.Fields{
		"host":    i.config.ReplsetInit.PrimaryAddr,
		"auth":    true,
		"replset": i.config.Replset,
	}).Info("Connected to MongoDB")

	err = i.initUsers(replsetAuthSession)
	if err != nil {
		return err
	}

	log.Info("Mongod replset initiator complete")
	return nil
}
