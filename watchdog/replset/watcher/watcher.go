package watcher

import (
	"sync"
	"time"

	"github.com/mesosphere/dcos-mongo/mongodb_tools/common"
	"github.com/mesosphere/dcos-mongo/mongodb_tools/watchdog/config"
	"github.com/mesosphere/dcos-mongo/mongodb_tools/watchdog/replset"
	log "github.com/sirupsen/logrus"
	rs_config "github.com/timvaillancourt/go-mongodb-replset/config"
	"gopkg.in/mgo.v2"
)

var (
	directReadPreference               = mgo.Monotonic
	replsetReadPreference              = mgo.PrimaryPreferred
	waitForMongodAvailableRetries uint = 10
)

type Watcher struct {
	config            *config.Config
	masterSession     *mgo.Session
	masterSessionLock sync.Mutex
	mongodAddQueue    chan []*replset.Mongod
	mongodRemoveQueue chan []*rs_config.Member
	replset           *replset.Replset
	state             *replset.State
	stop              chan bool
}

func New(rs *replset.Replset, config *config.Config, stop chan bool) *Watcher {
	return &Watcher{
		config:            config,
		replset:           rs,
		mongodAddQueue:    make(chan []*replset.Mongod),
		mongodRemoveQueue: make(chan []*rs_config.Member),
		stop:              stop,
	}
}

func (rw *Watcher) doStop() bool {
	if common.DoStop(&rw.stop) == true {
		log.WithFields(log.Fields{
			"replset": rw.replset.Name,
		}).Info("Stopping watcher for replset")
		return true
	}
	return false
}

func (rw *Watcher) getReplsetSession() (*mgo.Session, error) {
	rw.masterSessionLock.Lock()
	defer rw.masterSessionLock.Unlock()

	if rw.masterSession == nil {
		dialInfo := rw.replset.GetReplsetDialInfo()
		session, err := mgo.DialWithInfo(dialInfo)
		if err != nil {
			return nil, err
		}
		session.SetMode(replsetReadPreference, true)
		rw.masterSession = session
	}

	return rw.masterSession.Copy(), nil
}

func (rw *Watcher) logState() {
	primary := rw.state.Status.Primary()
	member := rw.replset.GetMember(primary.Name)
	log.WithFields(log.Fields{
		"replset":    rw.replset.Name,
		"host":       primary.Name,
		"task":       member.Task.Name(),
		"task_state": member.Task.State(),
	}).Info("Replset Primary")
	for _, secondary := range rw.state.Status.Secondaries() {
		member = rw.replset.GetMember(secondary.Name)
		log.WithFields(log.Fields{
			"replset":    rw.replset.Name,
			"host":       secondary.Name,
			"task":       member.Task.Name(),
			"task_state": member.Task.State(),
		}).Info("Replset Secondary")
	}
}

func (rw *Watcher) getMongodsNotInReplsetConfig() []*replset.Mongod {
	notInReplset := make([]*replset.Mongod, 0)
	if rw.state != nil && rw.state.Config != nil {
		for _, member := range rw.replset.GetMembers() {
			cnfMember := rw.state.Config.GetMember(member.Name())
			if cnfMember == nil {
				notInReplset = append(notInReplset, member)
			}
		}
	}
	return notInReplset
}

func (rw *Watcher) getOrphanedMembersFromReplsetConfig() []*rs_config.Member {
	orphanedMembers := make([]*rs_config.Member, 0)
	if rw.state != nil && rw.state.Config != nil {
		for _, member := range rw.state.Config.Members {
			if rw.replset.HasMember(member.Host) != true {
				orphanedMembers = append(orphanedMembers, member)
			}
		}
	}
	return orphanedMembers
}

func (rw *Watcher) waitForAvailable(mongod *replset.Mongod) error {
	session, err := common.WaitForSession(
		mongod.DBConfig(),
		waitForMongodAvailableRetries,
		rw.config.ReplsetPoll,
	)
	if err != nil {
		return err
	}
	session.Close()
	return nil
}

func (rw *Watcher) replsetConfigAdder(add <-chan []*replset.Mongod) {
	for members := range add {
		mongods := make([]*replset.Mongod, 0)
		for _, mongod := range members {
			err := rw.waitForAvailable(mongod)
			if err != nil {
				log.WithFields(log.Fields{
					"host":    mongod.Name(),
					"retries": waitForMongodAvailableRetries,
				}).Error(err)
				continue
			}
			log.WithFields(log.Fields{
				"replset": rw.replset.Name,
				"host":    mongod.Name(),
			}).Info("Mongod not present in replset config, adding it to replset")
			mongods = append(mongods, mongod)
		}
		rw.state.AddConfigMembers(mongods)
	}
}

func (rw *Watcher) replsetConfigRemover(remove <-chan []*rs_config.Member) {
	for members := range remove {
		if rw.state == nil {
			continue
		}
		rw.state.RemoveConfigMembers(members)
	}
}

func (rw *Watcher) Run() {
	log.WithFields(log.Fields{
		"replset": rw.replset.Name,
	}).Info("Watching replset")

	go rw.replsetConfigAdder(rw.mongodAddQueue)
	go rw.replsetConfigRemover(rw.mongodRemoveQueue)

	for !rw.doStop() {
		if rw.state == nil {
			session, err := rw.getReplsetSession()
			if err != nil {
				log.WithFields(log.Fields{
					"replset": rw.replset.Name,
					"error":   err,
				}).Error("Error getting mongodb session to replset")
			} else {
				rw.state = replset.NewState(session, rw.replset.Name)
				go rw.state.StartFetcher(rw.stop, rw.config.ReplsetPoll)
			}
		}

		if rw.state != nil && rw.state.Status != nil {
			rw.mongodAddQueue <- rw.getMongodsNotInReplsetConfig()
			rw.mongodRemoveQueue <- rw.getOrphanedMembersFromReplsetConfig()
			rw.logState()
		}

		time.Sleep(rw.config.ReplsetPoll)
	}
}
