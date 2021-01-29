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
	"github.com/dapr/dapr/dapr-agent/pkg/model"
	"github.com/gogo/protobuf/types"
)

type LocalhostValue string
type WildcardValue string
type DNSLookupFamilyValue string

const (
	LocalhostIPv4       LocalhostValue       = "127.0.0.1"
	LocalhostIPv6       LocalhostValue       = "::1"
	WildcardIPv4        WildcardValue        = "0.0.0.0"
	WildcardIPv6        WildcardValue        = "::"
	DNSLookupFamilyIPv4 DNSLookupFamilyValue = "V4_ONLY"
	DNSLookupFamilyIPv6 DNSLookupFamilyValue = "AUTO"
)

func ConnectTimeout(value *types.Duration) Instance {
	return newDurationOption("connect_timeout", value)
}

func Cluster(value string) Instance {
	return newOption("cluster", value)
}

func NodeID(value string) Instance {
	return newOption("nodeID", value)
}

func Region(value string) Instance {
	return newOptionOrSkipIfZero("region", value)
}

func Zone(value string) Instance {
	return newOptionOrSkipIfZero("zone", value)
}

func SubZone(value string) Instance {
	return newOptionOrSkipIfZero("sub_zone", value)
}

func NodeMetadata(meta *model.NodeMetadata, rawMeta map[string]interface{}) Instance {
	return newOptionOrSkipIfZero("meta_json_str", meta).withConvert(nodeMetadataConverter(meta, rawMeta))
}

func DiscoveryAddress(value string) Instance {
	return newOption("discovery_address", value)
}

func DiscoveryChannelCreds(value string) Instance {
	return newOption("discovery_channel_creds", value)
}

func EnabledFeatures(value []string) Instance {
	return newOption("server_features_str", value).withConvert(strSliceConverter(value))
}
func Localhost(value LocalhostValue) Instance {
	return newOption("localhost", value)
}

func Wildcard(value WildcardValue) Instance {
	return newOption("wildcard", value)
}

func DNSLookupFamily(value DNSLookupFamilyValue) Instance {
	return newOption("dns_lookup_family", value)
}

func PilotGRPCAddress(value string) Instance {
	return newOptionOrSkipIfZero("pilot_grpc_address", value).withConvert(addressConverter(value))
}

func DiscoveryHost(value string) Instance {
	return newOption("discovery_host", value)
}
