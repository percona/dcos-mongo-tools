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
	gotesting "testing"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/stretchr/testify/assert"
)

func TestAPIGetBaseURL(t *gotesting.T) {
	expected := DefaultHTTPHostPrefix + "." + common.DefaultFrameworkName + "." + DefaultHTTPHostSuffix
	assert.Equal(t, expected, testAPI.getBaseURL(), "api.getBaseURL() is incorrect")
}
