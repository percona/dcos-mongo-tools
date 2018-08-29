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

package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/percona/dcos-mongo-tools/executor/metrics"
	"github.com/percona/dcos-mongo-tools/executor/mongodb"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
	"github.com/percona/dcos-mongo-tools/internal"
	"github.com/percona/dcos-mongo-tools/internal/db"
)

const (
	DefaultDelayBackgroundJob = "15s"
	DefaultConnectRetrySleep  = "5s"
)

type NodeType string

const (
	NodeTypeMongod NodeType = "mongod"
	NodeTypeMongos NodeType = "mongos"
)

func (t NodeType) String() string {
	return string(t)
}

type Config struct {
	DB                 *db.Config
	MongoDB            *mongodb.Config
	PMM                *pmm.Config
	Metrics            *metrics.Config
	NodeType           NodeType
	FrameworkName      string
	DelayBackgroundJob time.Duration
	ConnectRetrySleep  time.Duration
	Verbose            bool
}

func MesosSandboxPathOrFallback(path string, fallback string) string {
	mesosSandbox := os.Getenv(internal.EnvMesosSandbox)
	if mesosSandbox != "" {
		if _, err := os.Stat(mesosSandbox); err == nil {
			return filepath.Join(mesosSandbox, path)
		}
	}
	return fallback
}
