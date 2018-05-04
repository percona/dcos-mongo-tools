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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	DefaultHostPrefix = "api"
	DefaultHostSuffix = "marathon.l4lb.thisdcos.directory"
	DefaultTimeout    = "5s"
)

type Config struct {
	HostPrefix string
	HostSuffix string
	Timeout    time.Duration
}

type Api struct {
	FrameworkName string
	config        *Config
	client        *http.Client
}

func New(frameworkName string, config *Config) *Api {
	return &Api{
		FrameworkName: frameworkName,
		config:        config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (a *Api) GetBaseUrl() string {
	return a.config.HostPrefix + "." + a.FrameworkName + "." + a.config.HostSuffix
}

func (a *Api) GetPodUrl() string {
	return "http://" + a.GetBaseUrl() + "/v1/pod"
}

func (a *Api) GetPods() (*Pods, error) {
	pods := &Pods{}
	err := a.Get(a.GetPodUrl(), pods)
	return pods, err
}

func (a *Api) GetPodTasks(podName string) ([]PodTask, error) {
	podUrl := a.GetPodUrl() + "/" + podName + "/info"
	var tasks []PodTask
	err := a.Get(podUrl, &tasks)
	return tasks, err
}

func (a *Api) GetEndpointsUrl() string {
	return "http://" + a.GetBaseUrl() + "/v1/endpoints"
}

func (a *Api) GetEndpoints() (*Endpoints, error) {
	endpoints := &Endpoints{}
	err := a.Get(a.GetEndpointsUrl(), endpoints)
	return endpoints, err
}

func (a *Api) GetEndpoint(endpointName string) (*Endpoint, error) {
	endpointUrl := a.GetEndpointsUrl() + "/" + endpointName
	endpoint := &Endpoint{}
	err := a.Get(endpointUrl, endpoint)
	return endpoint, err
}

func (a *Api) Get(url string, out interface{}) error {
	resp, err := a.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}
