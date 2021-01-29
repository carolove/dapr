// Copyright Dapr Authors
//
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

package config

import (
	"io/ioutil"
	"os"

	"github.com/dapr/dapr/dapr-agent/pkg/bootstrap"
	"github.com/dapr/dapr/dapr-agent/pkg/model"
	"istio.io/api/annotation"
	"istio.io/istio/pkg/config/constants"
)

// 这个地方应该construct dapr runtime config
// 然后在其他地方包裹这个config
// return proxyConfig and trustDomain
func ConstructProxyConfig(meta *model.NodeMetadata) (*model.NodeMetadata, error) {
	annotations, err := readPodAnnotations()
	if err != nil {
		if os.IsNotExist(err) {
			log.Debugf("failed to read pod annotations: %v", err)
		} else {
			log.Warnf("failed to read pod annotations: %v", err)
		}
	}
	meshConfig := model.DefaultMetaConfig(meta)

	return applyAnnotations(meshConfig, annotations), nil
}

func readPodAnnotations() (map[string]string, error) {
	b, err := ioutil.ReadFile(constants.PodInfoAnnotationsPath)
	if err != nil {
		return nil, err
	}
	return bootstrap.ParseDownwardAPI(string(b))
}

// Apply any overrides to proxy config from annotations
func applyAnnotations(config *model.NodeMetadata, annos map[string]string) *model.NodeMetadata {
	if v, f := annos[annotation.SidecarDiscoveryAddress.Name]; f {
		config.DiscoveryAddress = v
	}
	return config
}
