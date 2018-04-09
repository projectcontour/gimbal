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

package translator

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

const (
	// GimbalLabelCluster is the key of the label that contains the cluster name
	GimbalLabelCluster = "gimbal.heptio.com/cluster"
	gimbalLabelService = "gimbal.heptio.com/service"
	// Services in Kubernetes are limited to 63 characters in length
	maxNameLength = 63
)

// QueueItem represents and item in the queue
type QueueItem interface {
	Handle() (interface{}, error)
	IsSystem() bool
}

// AddGimbalLabels returns a new set of labels that includes the incoming set of
// labels, plus gimbal-specific ones.
func AddGimbalLabels(clustername, namespace, name string, existingLabels map[string]string) map[string]string {
	gimbalLabels := map[string]string{
		GimbalLabelCluster: clustername,
		gimbalLabelService: name,
	}
	if existingLabels == nil {
		return gimbalLabels
	}
	// Set gimbal labels on the existing labels map
	for k, v := range gimbalLabels {
		existingLabels[k] = v
	}
	return existingLabels
}

// GetFormattedName take the names of a service and formats to truncate if needed
func GetFormattedName(name, cluster string) string {
	return hashname(maxNameLength, name, cluster)
}

// hashname takes a lenth l and a varargs of strings s and returns a string whose length
// which does not exceed l. Internally s is joined with strings.Join(s, "/"). If the
// combined length exceeds l then hashname truncates each element in s, starting from the
// end using a hash derived from the contents of s (not the current element). This process
// continues until the length of s does not exceed l, or all elements have been truncated.
// In which case, the entire string is replaced with a hash not exceeding the length of l.
func hashname(l int, s ...string) string {
	const shorthash = 6 // the length of the shorthash

	r := strings.Join(s, "-")
	if l > len(r) {
		// we're under the limit, nothing to do
		return r
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(r)))
	for n := len(s) - 1; n >= 0; n-- {
		s[n] = truncate(l/len(s), s[n], hash[:shorthash])
		r = strings.Join(s, "-")
		if l > len(r) {
			return r
		}
	}
	// truncated everything, but we're still too long
	// just return the hash truncated to l.
	return hash[:min(len(hash), l)]
}

// truncate truncates s to l length by replacing the
// end of s with -suffix.
func truncate(l int, s, suffix string) string {
	if l >= len(s) {
		// under the limit, nothing to do
		return s
	}
	if l <= len(suffix) {
		// easy case, just return the start of the suffix
		return suffix[:min(l, len(suffix))]
	}
	return s[:l-len(suffix)-1] + "-" + suffix
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
