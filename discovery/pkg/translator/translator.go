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
	// GimbalLabelBackend is the key of the label that contains the cluster name
	GimbalLabelBackend          = "gimbal.heptio.com/backend"
	gimbalLabelService          = "gimbal.heptio.com/service"
	maxKubernetesDNSLabelLength = 63
)

// AddGimbalLabels returns a new set of labels that includes the incoming set of
// labels, plus gimbal-specific ones.
func AddGimbalLabels(backendname, name string, existingLabels map[string]string) map[string]string {
	gimbalLabels := map[string]string{
		GimbalLabelBackend: ShortenKubernetesLabelValue(backendname),
		gimbalLabelService: ShortenKubernetesLabelValue(name),
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

// BuildDiscoveredName returns the discovered name of the service in a given
// cluster. If the name is longer than the Kubernetes DNS_LABEL maximum
// character limit, the name is shortened.
func BuildDiscoveredName(backendName, serviceName string) string {
	return hashname(maxKubernetesDNSLabelLength, backendName, serviceName)
}

// ShortenKubernetesLabelValue ensures that the given string's length does not
// exceed the character limit imposed by Kubernetes. If it does, the label is
// shortened.
func ShortenKubernetesLabelValue(value string) string {
	return hashname(maxKubernetesDNSLabelLength, value)
}

// hashname takes a length l and a varargs of strings s and returns a string
// whose length which does not exceed l. Internally s is joined with
// strings.Join(s, "-"). If the combined length exceeds l then hashname
// truncates each element in s, starting from the end using a hash derived from
// the element. This process continues until the length of s does not exceed l.
func hashname(l int, s ...string) string {
	const shorthash = 6 // the length of the shorthash

	r := strings.Join(s, "-")
	if l > len(r) {
		// we're under the limit, nothing to do
		return r
	}
	for i := len(s) - 1; i >= 0; i-- {
		hash := fmt.Sprintf("%x", sha256.Sum256([]byte(s[i])))
		s[i] = truncate(l/len(s), s[i], hash[:shorthash])
		r = strings.Join(s, "-")
		if l > len(r) {
			// r is already short enough.
			break
		}
	}
	return r
}

// truncate truncates s to l length by replacing the
// end of s with suffix.
func truncate(l int, s, suffix string) string {
	if l >= len(s) {
		// under the limit, nothing to do
		return s
	}
	if l <= len(suffix) {
		// easy case, just return the start of the suffix
		return suffix[:min(l, len(suffix))]
	}
	return s[:l-len(suffix)] + suffix
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
