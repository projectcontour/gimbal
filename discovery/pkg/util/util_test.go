// Copyright © 2018 Heptio
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslateService(t *testing.T) {
	tests := []struct {
		name        string
		backendName string
		expected    bool
	}{
		{
			name:        "empty string",
			backendName: "",
			expected:    true,
		},
		{
			name:        "simple",
			backendName: "mycluster",
			expected:    false,
		},
		{
			name:        "hyphen",
			backendName: "my-cluster",
			expected:    false,
		},
		{
			name:        "underscore",
			backendName: "my_cluster",
			expected:    true,
		},
		{
			name:        "multiple underscores",
			backendName: "my----cluster",
			expected:    false,
		},
		{
			name:        "can't start with underscores",
			backendName: "-mycluster",
			expected:    true,
		},
		{
			name:        "can't end with underscores",
			backendName: "mycluster-",
			expected:    true,
		},
		{
			name:        "special chars",
			backendName: "!@!mycl^%$uster**",
			expected:    true,
		},
		{
			name:        "special chars with hyphen & underscore",
			backendName: "!@!my-cl^%$ust_er**",
			expected:    true,
		},
		{
			name:        "whitespace",
			backendName: "  my cluster  ",
			expected:    true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsInvalidBackendName(tc.backendName)
			assert.EqualValues(t, tc.expected, got)
		})
	}
}
