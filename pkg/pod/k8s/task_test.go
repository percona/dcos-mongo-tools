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

package k8s

import (
	"testing"

	"github.com/percona/mongodb-orchestration-tools/pkg/pod"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInternalPodK8STask(t *testing.T) {
	assert.Implements(t, (*pod.Task)(nil), &Task{})

	task := NewTask(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.Name(),
		},
	}, "mongodb")

	assert.NotNil(t, task)
	assert.Equal(t, t.Name(), task.Name())
}
