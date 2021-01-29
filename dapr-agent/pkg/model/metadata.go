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

package model

import (
	"path"
	"strings"

	"github.com/dapr/dapr/dapr-agent/pkg/platform"
	"istio.io/istio/pkg/config/constants"
)

const (
	serviceNodeSeparator = "~"
)

// StringList is a list that will be marshaled to a comma separate string in Json
type StringList []string

func (l StringList) MarshalJSON() ([]byte, error) {
	if l == nil {
		return nil, nil
	}
	return []byte(`"` + strings.Join(l, ",") + `"`), nil
}

func (l *StringList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == `""` {
		*l = []string{}
	} else {
		*l = strings.Split(string(data[1:len(data)-1]), ",")
	}
	return nil
}

// DefaultMetaConfig returns the default mesh config.
// This is merged with values from the mesh config map.
func DefaultMetaConfig(meta *NodeMetadata) *NodeMetadata {

	// Defaults matching the standard install
	// order matches the generated mesh config.
	if meta.ConfigPath == "" {
		meta.ConfigPath = "/etc/dapr/runtime"
	}
	if meta.ServiceCluster == "" {
		meta.ServiceCluster = constants.ServiceClusterName
	}
	if meta.DiscoveryAddress == "" {
		meta.DiscoveryAddress = "xds-server.dapr-system.svc:15010"
	}
	if meta.BinaryPath == "" {
		meta.BinaryPath = "/usr/local/bin/daprd"
	}
	if meta.ID == "" {
		meta.ID = meta.PodName + "." + meta.PodNamespace
	}
	if meta.ClusterID == "" {
		meta.ClusterID = meta.PodName + "." + meta.PodNamespace
	}
	if meta.DNSDomain == "" {
		meta.DNSDomain = getDNSDomain(meta.PodNamespace, meta.DNSDomain)
	}
	return meta
}

type NodeMetadata struct {
	// from bootstrap

	ID           string
	Node         string
	PodName      string
	PodNamespace string
	PlatEnv      platform.Environment
	LocalEnv     []string

	// from proxy config
	LogLevel          string
	ComponentLogLevel string
	Sidecar           bool
	ExtraArgs         []string

	// dapr runtime config

	// dapr params
	Mode                 string
	EnableMesh           bool
	DaprHTTPPort         string
	DaprAPIGRPCPort      string
	DaprInternalGRPCPort string
	AppPort              string
	ProfilePort          string
	AppProtocol          string
	ComponentsPath       string
	// config string
	AppID                    string
	ControlPlaneAddress      string
	SentryAddress            string
	PlacementServiceHostAddr string
	AllowedOrigins           string
	EnableProfiling          bool
	RuntimeVersion           bool
	AppMaxConcurrency        int
	EnableMTLS               bool
	AppSSL                   bool
	DaprHTTPMaxRequestSize   int

	// InstanceName is the short name for the workload instance (ex: pod name)
	// replaces POD_NAME
	InstanceName string `json:"NAME,omitempty"`

	// WorkloadName specifies the name of the workload represented by this node.
	WorkloadName string `json:"WORKLOAD_NAME,omitempty"`

	// Owner specifies the workload owner (opaque string). Typically, this is the owning controller of
	// of the workload instance (ex: k8s deployment for a k8s pod).
	Owner string `json:"OWNER,omitempty"`

	// PlatformMetadata contains any platform specific metadata
	PlatformMetadata map[string]string `json:"PLATFORM_METADATA,omitempty"`

	// Namespace is the namespace in which the workload instance is running.
	Namespace string `json:"NAMESPACE,omitempty"`
	// ServiceAccount specifies the service account which is running the workload.
	ServiceAccount string `json:"SERVICE_ACCOUNT,omitempty"`
	// MeshID specifies the mesh ID environment variable.
	MeshID string `json:"MESH_ID,omitempty"`
	// ClusterID defines the cluster the node belongs to.
	ClusterID string `json:"CLUSTER_ID,omitempty"`

	// DNSDomain defines the DNS domain suffix for short hostnames (e.g.
	// "default.svc.cluster.local")
	DNSDomain string

	// Service cluster defines the name for the `service_cluster` that is
	// shared by all dapr runtime instances. This setting corresponds to
	// `--service-cluster` flag in dapr runtime.  In a typical dapr runtime deployment, the
	// `service-cluster` flag is used to identify the caller, for
	// source-based routing scenarios.
	ServiceCluster string `protobuf:"bytes,3,opt,name=service_cluster,json=serviceCluster,proto3" json:"serviceCluster,omitempty"`
	// Address of the discovery service exposing xDS with mTLS connection.
	// The inject configuration may override this value.
	DiscoveryAddress string `protobuf:"bytes,6,opt,name=discovery_address,json=discoveryAddress,proto3" json:"discoveryAddress,omitempty"`
	// File path of custom proxy configuration, currently used by proxies
	// in front of Mixer and Pilot.
	CustomConfigFile string `protobuf:"bytes,14,opt,name=custom_config_file,json=customConfigFile,proto3" json:"customConfigFile,omitempty"`
	// Path to the proxy binary
	BinaryPath string `protobuf:"bytes,2,opt,name=binary_path,json=binaryPath,proto3" json:"binaryPath,omitempty"`
	// Path to the generated configuration file directory.
	// Proxy agent generates the actual configuration and stores it in this directory.
	ConfigPath string `protobuf:"bytes,1,opt,name=config_path,json=configPath,proto3" json:"configPath,omitempty"`

	// Labels specifies the set of workload instance (ex: k8s pod) labels associated with this node.
	Labels map[string]string `json:"LABELS,omitempty"`

	// InstanceIPs is the set of IPs attached to this proxy
	InstanceIPs StringList `json:"INSTANCE_IPS,omitempty"`
}

func getDNSDomain(podNamespace, domain string) string {
	if len(domain) == 0 {
		domain = podNamespace + ".svc." + constants.DefaultKubernetesDomain
	}
	return domain
}

func ConfigFile(config string, epoch int) string {
	return path.Join(config, "grpc_bootstrap.json")
}

// ServiceNode encodes the proxy node attributes into a URI-acceptable string
func (node *NodeMetadata) ServiceNode() string {
	ip := ""
	if len(node.InstanceIPs) > 0 {
		ip = node.InstanceIPs[0]
	}
	return strings.Join([]string{
		"sidecar", ip, node.ID, node.DNSDomain,
	}, serviceNodeSeparator)
}
