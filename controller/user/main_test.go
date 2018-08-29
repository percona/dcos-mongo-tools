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

package user

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	gotesting "testing"
	"time"

	"github.com/percona/dcos-mongo-tools/controller"
	"github.com/percona/dcos-mongo-tools/internal"
	"github.com/percona/dcos-mongo-tools/internal/db"
	"github.com/percona/dcos-mongo-tools/internal/logger"
	"github.com/percona/dcos-mongo-tools/internal/testing"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	testDirRelPath     = "./json/testdata"
	testBase64JSONFile = "test-user.json.base64"
)

var (
	testSession     *mgo.Session
	testController  *Controller
	testLogBuffer   = new(bytes.Buffer)
	testSystemUsers = []*mgo.User{
		{Username: "testAdmin", Password: "123456", Roles: []mgo.Role{"root"}},
	}
	testControllerConfig = &controller.Config{
		SSL: &db.SSLConfig{},
		User: &controller.ConfigUser{
			Database:        SystemUserDatabase,
			File:            internal.RelPathToAbs(filepath.Join(testDirRelPath, testBase64JSONFile)),
			Username:        "prodapp",
			EndpointName:    internal.DefaultMongoDBMongodEndpointName,
			MaxConnectTries: 1,
			RetrySleep:      time.Second,
		},
		FrameworkName:     internal.DefaultFrameworkName,
		Replset:           testing.MongodbReplsetName,
		UserAdminUser:     testing.MongodbAdminUser,
		UserAdminPassword: testing.MongodbAdminPassword,
	}
)

func checkUserExists(session *mgo.Session, user, db string) error {
	resp := struct {
		Username string `bson:"user"`
		Database string `bson:"db"`
	}{}
	err := session.DB(testControllerConfig.User.Database).C("system.users").Find(bson.M{
		"user": user,
		"db":   db,
	}).One(&resp)
	if err != nil {
		return err
	}
	if resp.Username != user || resp.Database != db {
		return errors.New("user does not match")
	}
	return nil
}

func TestMain(m *gotesting.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter("test"), testLogBuffer)

	if testing.Enabled() {
		var err error
		testSession, err = testing.GetSession(testing.MongodbPrimaryPort)
		if err != nil {
			panic(err)
		}
	}

	exit := m.Run()

	if testSession != nil {
		testSession.Close()
	}
	if testController != nil {
		testController.Close()
	}
	os.Exit(exit)
}
