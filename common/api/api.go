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

package api

import (
	"time"
)

type ApiScheme string

const (
	ApiSchemePlain  ApiScheme = "http://"
	ApiSchemeSecure ApiScheme = "https://"
)

func (s ApiScheme) String() string {
	return string(s)
}

type Config struct {
	HostPrefix string
	HostSuffix string
	Timeout    time.Duration
	Secure     bool
}

type Api interface {
	GetBaseUrl() string
	GetPodUrl() string
	GetPods() (*Pods, error)
	GetPodTasks(podName string) ([]PodTask, error)
	GetEndpointsUrl() string
	GetEndpoints() (*Endpoints, error)
	GetEndpoint(endpointName string) (*Endpoint, error)
}
