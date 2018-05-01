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

package metrics

import (
	"time"

	mgostatsd "github.com/scullxbones/mgo-statsd"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

type Metrics struct {
	config  *Config
	running bool
	session *mgo.Session
}

func New(config *Config, session *mgo.Session) *Metrics {
	return &Metrics{
		config:  config,
		session: session,
	}
}

func (m *Metrics) Name() string {
	return "DC/OS Metrics"
}

func (m *Metrics) DoRun() bool {
	return m.config.Enabled
}

func (m *Metrics) IsRunning() bool {
	return m.running
}

func (m *Metrics) Run(quit *chan bool) error {
	if m.DoRun() == false {
		log.Warn("DC/OS Metrics client executor disabled! Skipping start")
		return nil
	}

	log.WithFields(log.Fields{
		"interval":    m.config.Interval,
		"statsd_host": m.config.StatsdHost,
		"statsd_port": m.config.StatsdPort,
	}).Info("Starting DC/OS Metrics pusher")

	ticker := time.NewTicker(m.config.Interval)
	statsdCnf := mgostatsd.Statsd{
		Host: m.config.StatsdHost,
		Port: m.config.StatsdPort,
	}

	m.running = true
	for {
		select {
		case <-ticker.C:
			status := mgostatsd.GetServerStatus(m.session)
			if status == nil {
				continue
			}

			log.WithFields(log.Fields{
				"statsd_host": m.config.StatsdHost,
				"statsd_port": m.config.StatsdPort,
			}).Info("Pushing DC/OS Metrics")

			err := mgostatsd.PushStats(statsdCnf, status, false)
			if err != nil {
				log.Errorf("DC/OS Metrics push error: %s", err)
			}
		case <-*quit:
			log.Info("Stopping DC/OS Metrics pusher")
			ticker.Stop()
			break
		}
	}

	m.running = false
	return nil
}
