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

package testing

import (
	"os"
	gotesting "testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

const (
	envEnableDBTests        = "ENABLE_MONGODB_TESTS"
	envMongoDBReplsetName   = "TEST_RS_NAME"
	envMongoDBPrimaryPort   = "TEST_PRIMARY_PORT"
	envMongoDBAdminUser     = "TEST_ADMIN_USER"
	envMongoDBAdminPassword = "TEST_ADMIN_PASSWORD"
)

var (
	enableDBTests        = os.Getenv(envEnableDBTests)
	MongodbReplsetName   = os.Getenv(envMongoDBReplsetName)
	MongodbPrimaryHost   = "127.0.0.1"
	MongodbPrimaryPort   = os.Getenv(envMongoDBPrimaryPort)
	MongodbAdminUser     = os.Getenv(envMongoDBAdminUser)
	MongodbAdminPassword = os.Getenv(envMongoDBAdminPassword)
	MongodbTimeout       = time.Duration(10) * time.Second
	dbSession            *mgo.Session
)

// Enabled returns a boolean reflecting whether testing against Mongodb should occur
func Enabled() bool {
	return enableDBTests == "true"
}

// getPrimaryDialInfo returns a *mgo.DialInfo configured for testing
func getDialInfo(t *gotesting.T, host, port string) *mgo.DialInfo {
	if Enabled() {
		assert.NotEmpty(t, port, "Port arguement is not set")
		assert.NotEmpty(t, host, "Host arguement is not set")
		assert.NotEmpty(t, MongodbReplsetName, "Replica set name env var is not set")
		assert.NotEmpty(t, MongodbAdminUser, "Admin user env var is not set")
		assert.NotEmpty(t, MongodbAdminPassword, "Admin password env var is not set")
		return &mgo.DialInfo{
			Addrs:          []string{host + ":" + port},
			Direct:         true,
			Timeout:        MongodbTimeout,
			Username:       MongodbAdminUser,
			Password:       MongodbAdminPassword,
			ReplicaSetName: MongodbReplsetName,
		}
	}
	return nil
}

// GetPrimarySession returns a *mgo.Session configured for testing against a MongoDB Primary
func GetPrimarySession(t *gotesting.T) *mgo.Session {
	if dbSession != nil {
		return dbSession
	}

	dialInfo := getDialInfo(t, MongodbPrimaryHost, MongodbPrimaryPort)
	assert.NotNil(t, dialInfo, "Could not build dial info for Primary")
	session, err := mgo.DialWithInfo(dialInfo)
	assert.NoError(t, err, "Database connection error")

	buildInfo, err := session.BuildInfo()
	assert.NoError(t, err, "Error getting database build info")

	t.Logf("Connected to PSMDB version: %s", buildInfo.Version)
	dbSession = session
	return dbSession
}

// Close closes the database session if it exists
func Close() {
	if dbSession != nil {
		dbSession.Close()
	}
}

// DoSkipTest handles the conditional skipping of tests, based on the output of .Enabled()
func DoSkipTest(t *gotesting.T) {
	if !Enabled() {
		t.Skipf("Skipping test, env var %s is not 'true'", envEnableDBTests)
	}
}
