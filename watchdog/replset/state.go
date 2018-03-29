package replset

import (
	"fmt"
	"sync"
	"time"

	"github.com/mesosphere/dcos-mongo/mongodb_tools/common"
	log "github.com/sirupsen/logrus"
	rs_config "github.com/timvaillancourt/go-mongodb-replset/config"
	rs_status "github.com/timvaillancourt/go-mongodb-replset/status"
	"gopkg.in/mgo.v2"
)

type State struct {
	sync.Mutex
	Replset       string
	Config        *rs_config.Config
	Status        *rs_status.Status
	configManager *rs_config.ConfigManager
	session       *mgo.Session
	doUpdate      bool
}

func NewState(session *mgo.Session, replset string) *State {
	return &State{
		Replset:       replset,
		configManager: rs_config.New(session),
		session:       session,
	}
}

func (s *State) fetchConfig() error {
	err := s.configManager.Load()
	if err != nil {
		return err
	}
	s.Config = s.configManager.Get()
	return nil
}

func (s *State) fetchStatus() error {
	status, err := rs_status.New(s.session)
	if err != nil {
		return err
	}
	s.Status = status
	return nil
}

func (s *State) Fetch() error {
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

func (s *State) updateConfig() error {
	if s.doUpdate == false {
		return nil
	}

	s.configManager.IncrVersion()
	config := s.configManager.Get()
	log.WithFields(log.Fields{
		"replset":        s.Replset,
		"config_version": config.Version,
	}).Info("Writing new replset config")
	fmt.Println(config.ToString())

	err := s.configManager.Save()
	if err != nil {
		return err
	}
	s.doUpdate = false

	return nil
}

func (s *State) AddConfigMembers(mongods []*Mongod) {
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
		member := rs_config.NewMember(mongod.Name())
		member.Tags = &rs_config.ReplsetTags{
			"dcosFramework": mongod.FrameworkName,
		}
		if mongod.IsBackupNode() {
			member.Hidden = true
			member.Priority = 0
			member.Tags = &rs_config.ReplsetTags{
				"backup":        "true",
				"dcosFramework": mongod.FrameworkName,
			}
			member.Votes = 0
		}
		s.configManager.AddMember(member)
		s.doUpdate = true
	}
	s.updateConfig()
}

func (s *State) RemoveConfigMembers(members []*rs_config.Member) {
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

func (s *State) doStopFetcher(stop chan bool) bool {
	return common.DoStop(&stop)
}

func (s *State) StartFetcher(stop chan bool, sleep time.Duration) {
	log.WithFields(log.Fields{
		"replset": s.Replset,
	}).Info("Started background replset state fetcher")
	for !s.doStopFetcher(stop) {
		s.Fetch()
		time.Sleep(sleep)
	}
}
