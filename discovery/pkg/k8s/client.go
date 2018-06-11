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

package k8s

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClient returns a Kubernetes client using the given config. If no config is
// provided, assumes it is running inside a Kubernetes cluster and uses the
// in-cluster config.
func NewClient(kubeCfgFile string, logger *logrus.Logger) (kubernetes.Interface, error) {
	config, err := buildConfig(kubeCfgFile, logger)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

// NewClientWithQPS returns a Kubernetes client using the given configuration
// and rate limiting parameters. If no config is provided, assumes it is running
// inside a Kubernetes cluster and uses the in-cluster config.
func NewClientWithQPS(kubeCfgFile string, logger *logrus.Logger, qps float32, burst int) (kubernetes.Interface, error) {
	config, err := buildConfig(kubeCfgFile, logger)
	if err != nil {
		return nil, err
	}
	config.QPS = qps
	config.Burst = burst
	return kubernetes.NewForConfig(config)
}

func buildConfig(kubeCfgFile string, logger *logrus.Logger) (*rest.Config, error) {
	if kubeCfgFile != "" {
		logger.Infof("Using OutOfCluster k8s config with kubeConfigFile: %s", kubeCfgFile)
		config, err := clientcmd.BuildConfigFromFlags("", kubeCfgFile)
		if err != nil {
			return nil, err
		}

		return config, nil
	}
	logger.Info("Using InCluster k8s config")
	return rest.InClusterConfig()
}
