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
	"strconv"
	"strings"
	gotesting "testing"

	testing "github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/percona/dcos-mongo-tools/watchdog/replset/fetcher"
	"github.com/stretchr/testify/assert"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
)

func TestNewState(t *gotesting.T) {
	testing.DoSkipTest(t)

	configManager := rsConfig.New(testDBSession)

	testState = NewState(nil, configManager, fetcher.New(testDBSession, configManager), testing.MongodbReplsetName)
	assert.Equal(t, testState.Replset, testing.MongodbReplsetName, "replset.NewState() returned State struct with incorrect 'Replset' name")
	assert.Nil(t, testState.session, "replset.NewState() returned State struct with a session other than nil")
	assert.False(t, testState.doUpdate, "replset.NewState() returned State struct with 'doUpdate' set to true")
}

func TestFetchConfig(t *gotesting.T) {
	testing.DoSkipTest(t)

	err := testState.fetchConfig()
	assert.NoError(t, err, ".fetchConfig() should not return an error")

	assert.NotNil(t, testState.Config, "testState.Config is nil")
	assert.Equal(t, testState.Config.Name, testing.MongodbReplsetName, "testState.Config.Name is incorrect")
	assert.NotZero(t, testState.Config.Members, "testState.Config.Members has no members")
}

func TestFetchStatus(t *gotesting.T) {
	testing.DoSkipTest(t)

	err := testState.fetchStatus()
	assert.NoError(t, err, ".fetchStatus() should not return an error")

	assert.NotNil(t, testState.Status, "testState.Status is nil")
	assert.Equal(t, testState.Status.Set, testing.MongodbReplsetName, "testState.Status.Set is incorrect")
	assert.NotZero(t, testState.Status.Members, "testState.Status.Members has no members")
}

func TestFetch(t *gotesting.T) {
	testing.DoSkipTest(t)

	assert.NoError(t, testState.Fetch(), "testState.Fetch() failed with error")
}

func TestRemoveAddConfigMembers(t *gotesting.T) {
	testing.DoSkipTest(t)

	memberCount := len(testState.Config.Members)
	removeMember := testState.Config.Members[len(testState.Config.Members)-1]
	testState.RemoveConfigMembers([]*rsConfig.Member{removeMember})
	assert.False(t, testState.doUpdate, "testState.doUpdate is true after testState.RemoveConfigMembers()")
	assert.Len(t, testState.Config.Members, memberCount-1, "testState.Config.Members count did not reduce")

	hostPort := strings.SplitN(removeMember.Host, ":", 2)
	port, _ := strconv.Atoi(hostPort[1])
	addMongod := &Mongod{
		Host:          hostPort[0],
		Port:          port,
		Replset:       testing.MongodbReplsetName,
		FrameworkName: "test",
		PodName:       "mongo",
	}
	testState.AddConfigMembers([]*Mongod{addMongod})
	assert.Falsef(t, testState.doUpdate, "testState.doUpdate is true after testState.AddConfigMembers()")
	assert.Len(t, testState.Config.Members, memberCount, "testState.Config.Members count did not increase")

	member := testState.Config.GetMember(removeMember.Host)
	assert.NotNil(t, member, "testState.Config.HasMember() returned no member")
	assert.True(t, member.Tags.HasMatch(frameworkTagName, "test"), "member has missing replica set tag")
}
