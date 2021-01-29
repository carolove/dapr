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

package option_test

import (
	"github.com/dapr/dapr/dapr-agent/pkg/model"
	"github.com/dapr/dapr/dapr-agent/pkg/option"
	"testing"

	. "github.com/onsi/gomega"
)

// nolint: lll
func TestOptions(t *testing.T) {
	cases := []struct {
		testName    string
		key         string
		option      option.Instance
		expectError bool
		expected    interface{}
	}{
		{
			testName: "cluster",
			key:      "cluster",
			option:   option.Cluster("fake"),
			expected: "fake",
		},
		{
			testName: "nodeID",
			key:      "nodeID",
			option:   option.NodeID("fake"),
			expected: "fake",
		},
		{
			testName: "region",
			key:      "region",
			option:   option.Region("fake"),
			expected: "fake",
		},
		{
			testName: "zone",
			key:      "zone",
			option:   option.Zone("fake"),
			expected: "fake",
		},
		{
			testName: "sub zone",
			key:      "sub_zone",
			option:   option.SubZone("fake"),
			expected: "fake",
		},
		{
			testName: "node metadata nil",
			key:      "meta_json_str",
			option:   option.NodeMetadata(nil, nil),
			expected: nil,
		},
		{
			testName: "node metadata",
			key:      "meta_json_str",
			option: option.NodeMetadata(&model.NodeMetadata{
				InstanceName: "fake",
			}, map[string]interface{}{
				"key": "value",
			}),
			expected: "{\"InstanceName\":\"fake\",\"key\":\"value\"}",
		},
		{
			testName: "discovery address",
			key:      "discovery_address",
			option:   option.DiscoveryAddress("fake"),
			expected: "fake",
		},
		{
			testName: "localhost v4",
			key:      "localhost",
			option:   option.Localhost(option.LocalhostIPv4),
			expected: option.LocalhostValue("127.0.0.1"),
		},
		{
			testName: "localhost v6",
			key:      "localhost",
			option:   option.Localhost(option.LocalhostIPv6),
			expected: option.LocalhostValue("::1"),
		},
		{
			testName: "wildcard v4",
			key:      "wildcard",
			option:   option.Wildcard(option.WildcardIPv4),
			expected: option.WildcardValue("0.0.0.0"),
		},
		{
			testName: "wildcard v6",
			key:      "wildcard",
			option:   option.Wildcard(option.WildcardIPv6),
			expected: option.WildcardValue("::"),
		},
		{
			testName: "dns lookup family v4",
			key:      "dns_lookup_family",
			option:   option.DNSLookupFamily(option.DNSLookupFamilyIPv4),
			expected: option.DNSLookupFamilyValue("V4_ONLY"),
		},
		{
			testName: "dns lookup family v6",
			key:      "dns_lookup_family",
			option:   option.DNSLookupFamily(option.DNSLookupFamilyIPv6),
			expected: option.DNSLookupFamilyValue("AUTO"),
		},
		{
			testName: "pilot grpc address empty",
			key:      "pilot_grpc_address",
			option:   option.PilotGRPCAddress(""),
			expected: nil,
		},
		{
			testName: "pilot grpc address ipv4",
			key:      "pilot_grpc_address",
			option:   option.PilotGRPCAddress("127.0.0.1:80"),
			expected: "{\"address\": \"127.0.0.1\", \"port_value\": 80}",
		},
		{
			testName: "pilot grpc address ipv6",
			key:      "pilot_grpc_address",
			option:   option.PilotGRPCAddress("[2001:db8::100]:80"),
			expected: "{\"address\": \"2001:db8::100\", \"port_value\": 80}",
		},
		{
			testName:    "pilot grpc address ipv6 missing brackets",
			key:         "pilot_grpc_address",
			option:      option.PilotGRPCAddress("2001:db8::100:80"),
			expectError: true,
		},
		{
			testName: "pilot grpc address host port",
			key:      "pilot_grpc_address",
			option:   option.PilotGRPCAddress("fake:80"),
			expected: "{\"address\": \"fake\", \"port_value\": 80}",
		},
		{
			testName: "pilot grpc address no port",
			key:      "pilot_grpc_address",
			option:   option.PilotGRPCAddress("fake:"),
			expected: "{\"address\": \"fake\", \"port_value\": }",
		},
		{
			testName:    "pilot grpc address missing port",
			key:         "pilot_grpc_address",
			option:      option.PilotGRPCAddress("127.0.0.1"),
			expectError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.testName, func(t *testing.T) {
			g := NewWithT(t)

			params, err := option.NewTemplateParams(c.option)
			if c.expectError {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				actual, ok := params[c.key]
				if c.expected == nil {
					g.Expect(ok).To(BeFalse())
				} else {
					g.Expect(ok).To(BeTrue())
					g.Expect(actual).To(Equal(c.expected))
				}
			}
		})
	}
}
