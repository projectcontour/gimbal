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

package openstack

import (
	"fmt"
	"math"
	"net/http"
	"net/http/httptrace"
	"time"

	localmetrics "github.com/heptio/gimbal/discovery/pkg/metrics"
	"github.com/sirupsen/logrus"
)

// LogRoundTripper satisfies the http.RoundTripper interface and is used to
// customize the default Gophercloud RoundTripper to allow for logging.
type LogRoundTripper struct {
	RoundTripper      http.RoundTripper
	numReauthAttempts int
	Log               *logrus.Logger
	Metrics           *localmetrics.DiscovererMetrics
	BackendName       string
}

// RoundTrip performs a round-trip HTTP request and logs relevant information about it.
func (lrt *LogRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	lrt.Log.Debugf("Request URL: %s", request.URL)

	start := time.Now()
	var latency time.Duration

	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			latency = time.Now().Sub(start)
			lrt.Log.Debug("-- API Latency: ", math.Floor(latency.Seconds()*1e3))
		},
	}
	request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace))

	response, err := lrt.RoundTripper.RoundTrip(request)
	if response == nil {
		return nil, err
	}

	if response.StatusCode == http.StatusUnauthorized {
		if lrt.numReauthAttempts == 3 {
			return response, fmt.Errorf("tried to re-authenticate 3 times with no success")
		}
		lrt.numReauthAttempts++
	}

	lrt.Log.Debugf("-- Response Status: %s", response.Status)

	lrt.Metrics.APILatencyMetric(request.URL.Path, latency)

	return response, nil
}
