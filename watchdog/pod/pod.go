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

package pod

type Pods []string

func (p Pods) HasPod(name string) bool {
	for _, pod := range p {
		if pod == name {
			return true
		}
	}
	return false
}

//// GetPodURL returns a string representing the full HTTP URI to the 'GET /<version>/pod' API call
//func (c *ClientHTTP) GetPodURL() string {
//	return c.scheme.String() + c.config.Host + "/" + APIVersion + "/pod"
//}
//
//// GetPods returns a slice of existing Pods in the DC/OS SDK
//func (c *ClientHTTP) GetPods() (*Pods, error) {
//	pods := &Pods{}
//	err := c.get(c.GetPodURL(), pods)
//	return pods, err
//}
//
//// GetPodTasks returns a slice of PodTask for a given DC/OS SDK Pod by name
//func (c *ClientHTTP) GetPodTasks(podName string) ([]PodTask, error) {
//	tasks := make([]PodTask, 0)
//
//	tasksHTTP := make([]*PodTaskHTTP, 0)
//	podURL := c.GetPodURL() + "/" + podName + "/info"
//	err := c.get(podURL, &tasksHTTP)
//	if err != nil {
//		return tasks, err
//	}
//
//	for _, task := range tasksHTTP {
//		tasks = append(tasks, task)
//	}
//	return tasks, nil
//}
