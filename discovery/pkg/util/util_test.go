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
		clusterName string
		expected    string
	}{
		{
			name:        "empty string",
			clusterName: "",
			expected:    "",
		},
		{
			name:        "simple",
			clusterName: "mycluster",
			expected:    "mycluster",
		},
		{
			name:        "special chars",
			clusterName: "!@!mycl^%$uster**",
			expected:    "mycluster",
		},
		{
			name:        "special chars with hyphen & underscore",
			clusterName: "!@!my-cl^%$ust_er**",
			expected:    "my-clust_er",
		},
		{
			name:        "whitespace",
			clusterName: "  my cluster  ",
			expected:    "mycluster",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := SanitizeClusterName(tc.clusterName)
			assert.EqualValues(t, tc.expected, got)
		})
	}
}
