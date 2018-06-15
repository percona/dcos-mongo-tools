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
	"fmt"
	"time"

	"github.com/percona/dcos-mongo-tools/watchdog/replset/fetcher"
	log "github.com/sirupsen/logrus"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
	rsStatus "github.com/timvaillancourt/go-mongodb-replset/status"
	"gopkg.in/mgo.v2"
)

const frameworkTagName = "dcosFramework"

type State struct {
	Replset       string
	config        *rsConfig.Config
	configManager rsConfig.Manager
	session       *mgo.Session
	status        *rsStatus.Status
	session       *mgo.Session
	//fetcher       fetcher.Fetcher
	doUpdate bool
}

func NewState(session *mgo.Session, configManager rsConfig.Manager, fetcher fetcher.Fetcher, replset string) *State {
	return &State{
		Replset:       replset,
		session:       session,
		configManager: configManager,
		fetcher:       fetcher,
	}
}

func (m *Manager) fetchConfig() error {
	config, err := s.fetcher.GetConfig()
	if err != nil {
		return err
	}
	s.Config = config
	return nil
}

func (m *Manager) fetchStatus() error {
	status, err := s.fetcher.GetStatus()
	if err != nil {
		return err
	}
	s.Status = status
	return nil
}

func (m *Manager) Fetch() error {
	s.Lock()
	defer s.Unlock()

	log.WithFields(log.Fields{
		"replset": s.Replset,
	}).Info("Updating replset config and status")

	err := s.fetchConfig()
	if err != nil {
		return err
	}

	return s.fetchStatus()
}

func (m *Manager) updateConfig() error {
	if s.doUpdate == false {
		return nil
	}

	s.configManager.IncrVersion()
	config := s.configManager.Get()
	log.WithFields(log.Fields{
		"replset":        s.Replset,
		"config_version": config.Version,
	}).Info("Writing new replset config")
	fmt.Println(config)

	err := s.configManager.Save()
	if err != nil {
		return err
	}
	s.doUpdate = false

	return nil
}

func (m *Manager) AddConfigMembers(mongods []*Mongod) {
	if len(mongods) == 0 {
		return
	}

	s.Lock()
	defer s.Unlock()

	err := s.fetchConfig()
	if err != nil {
		log.Errorf("Error fetching config while adding members: '%s'", err.Error())
		return
	}

	for _, mongod := range mongods {
		member := rsConfig.NewMember(mongod.Name())
		member.Tags = &rsConfig.ReplsetTags{
			frameworkTagName: mongod.FrameworkName,
		}
		if mongod.IsBackupNode() {
			member.Hidden = true
			member.Priority = 0
			member.Tags = &rsConfig.ReplsetTags{
				"backup":         "true",
				frameworkTagName: mongod.FrameworkName,
			}
			member.Votes = 0
		}
		s.configManager.AddMember(member)
		s.doUpdate = true
	}
	s.updateConfig()
}

func (m *Manager) RemoveConfigMembers(members []*rsConfig.Member) {
	if len(members) == 0 {
		return
	}

	s.Lock()
	defer s.Unlock()

	err := s.fetchConfig()
	if err != nil {
		log.Errorf("Error fetching config while removing members: '%s'", err.Error())
		return
	}

	for _, member := range members {
		s.configManager.RemoveMember(member)
		s.doUpdate = true
	}
	s.updateConfig()
}

func (m *Manager) StartFetcher(stop *chan bool, interval time.Duration) {
	log.WithFields(log.Fields{
		"replset":  s.Replset,
		"interval": interval,
	}).Info("Started background replset state fetcher")

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := s.Fetch()
			if err != nil {
				log.WithFields(log.Fields{
					"replset": s.Replset,
				}).Errorf("Error fetching replset state: %s", err.Error())
			}
		case <-*stop:
			ticker.Stop()
			break
		}
	}
}
