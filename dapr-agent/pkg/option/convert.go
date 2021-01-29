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

package option

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/dapr/dapr/dapr-agent/pkg/model"
	"github.com/gogo/protobuf/types"
)

func nodeMetadataConverter(metadata *model.NodeMetadata, rawMeta map[string]interface{}) convertFunc {
	return func(*instance) (interface{}, error) {
		marshalString, err := marshalMetadata(metadata, rawMeta)
		if err != nil {
			return "", err
		}
		return marshalString, nil
	}
}

func addressConverter(addr string) convertFunc {
	return func(o *instance) (interface{}, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("unable to parse %s address %q: %v", o.name, addr, err)
		}
		if host == "$(HOST_IP)" {
			// Replace host with HOST_IP env var if it is "$(HOST_IP)".
			// This is to support some tracer setting (Datadog, Zipkin), where "$(HOST_IP)"" is used for address.
			// Tracer address used to be specified within proxy container params, and thus could be interpreted with pod HOST_IP env var.
			// Now tracer config is passed in with mesh config volumn at gateway, k8s env var interpretation does not work.
			// This is to achieve the same interpretation as k8s.
			hostIPEnv := os.Getenv("HOST_IP")
			if hostIPEnv != "" {
				host = hostIPEnv
			}
		}

		return fmt.Sprintf("{\"address\": \"%s\", \"port_value\": %s}", host, port), nil
	}
}

func durationConverter(value *types.Duration) convertFunc {
	return func(*instance) (interface{}, error) {
		return value.String(), nil
	}
}

func convertToJSON(v interface{}) string {
	if v == nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		log.Error(err.Error())
		return ""
	}
	return string(b)
}

// marshalMetadata combines type metadata and untyped metadata and marshals to json
// This allows passing arbitrary metadata to dapr runtime, while still supported typed metadata for known types
func marshalMetadata(metadata *model.NodeMetadata, rawMeta map[string]interface{}) (string, error) {
	b, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	var output map[string]interface{}
	if err := json.Unmarshal(b, &output); err != nil {
		return "", err
	}
	// Add all untyped metadata
	for k, v := range rawMeta {
		// Do not override fields, as we may have made modifications to the type metadata
		// This means we will only add "unknown" fields here
		if _, f := output[k]; !f {
			output[k] = v
		}
	}
	res, err := json.Marshal(output)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func strSliceConverter(strs []string) convertFunc {
	return func(*instance) (interface{}, error) {
		var rets []string
		for _, s := range strs {
			rets = append(rets, fmt.Sprintf(`"%s"`, s))
		}
		return "[" + strings.Join(rets, ",") + "]", nil
	}
}
