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
	"sync"
	"time"

	"github.com/percona/dcos-mongo-tools/watchdog/config"
	"gopkg.in/mgo.v2"
)

type Replset struct {
	sync.Mutex
	config      *config.Config
	Members     map[string]*Member
	Name        string
	LastUpdated time.Time
}

func New(config *config.Config, name string) *Replset {
	return &Replset{
		config:  config,
		Members: make(map[string]*Member),
		Name:    name,
	}
}

func (r *Replset) UpdateMember(mongod *Member) {
	r.Lock()
	defer r.Unlock()

	r.Members[mongod.Name()] = mongod
	r.LastUpdated = time.Now()
}

func (r *Replset) RemoveMember(mongod *Member) {
	r.Lock()
	defer r.Unlock()

	delete(r.Members, mongod.Name())
}

func (r *Replset) GetMember(name string) *Member {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.Members[name]; ok {
		return r.Members[name]
	}
	return nil
}

func (r *Replset) HasMember(name string) bool {
	return r.GetMember(name) != nil
}

func (r *Replset) GetMembers() map[string]*Member {
	r.Lock()
	defer r.Unlock()

	return r.Members
}

func (r *Replset) GetReplsetDialInfo() *mgo.DialInfo {
	di := &mgo.DialInfo{
		Direct:         false,
		FailFast:       true,
		ReplicaSetName: r.Name,
		Timeout:        r.config.ReplsetTimeout,
	}
	for _, member := range r.GetMembers() {
		di.Addrs = append(di.Addrs, member.Name())
	}
	if r.config.Username != "" && r.config.Password != "" {
		di.Username = r.config.Username
		di.Password = r.config.Password
	}
	return di
}
