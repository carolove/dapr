package bootstrap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/dapr/dapr/dapr-agent/pkg/model"
	"github.com/dapr/dapr/dapr-agent/pkg/platform"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

func TestGolden(t *testing.T) {

	out := "/tmp"

	cases := []struct {
		base          string
		envVars       map[string]string
		annotations   map[string]string
		checkLocality bool
		platformMeta  map[string]string
		setup         func()
		teardown      func()
	}{
		{
			base: "default",
			envVars: map[string]string{
				"ISTIO_META_ISTIO_PROXY_SHA":   "istio-proxy:sha",
				"ISTIO_META_INTERCEPTION_MODE": "REDIRECT",
				"ISTIO_META_ISTIO_VERSION":     "release-3.1",
				"ISTIO_META_POD_NAME":          "svc-0-0-0-6944fb884d-4pgx8",
				"POD_NAME":                     "svc-0-0-0-6944fb884d-4pgx8",
				"POD_NAMESPACE":                "test",
				"INSTANCE_IP":                  "10.10.10.1",
				"ISTIO_METAJSON_LABELS":        `{"version": "v1alpha1", "app": "test", "istio-locality":"regionA.zoneB.sub_zoneC"}`,
			},
			annotations: map[string]string{
				"istio.io/insecurepath": "{\"paths\":[\"/metrics\",\"/live\"]}",
			},
			checkLocality: true,
		},
	}

	for _, c := range cases {
		t.Run("Bootstrap-"+c.base, func(t *testing.T) {
			if c.setup != nil {
				c.setup()
			}
			if c.teardown != nil {
				defer c.teardown()
			}

			proxyConfig, err := loadProxyConfig(c.base, out, t)
			if err != nil {
				t.Fatalf("unable to load proxy config: %s\n%v", c.base, err)
			}

			_, localEnv := createEnv(t, map[string]string{}, c.annotations)
			for k, v := range c.envVars {
				localEnv = append(localEnv, k+"="+v)
			}
			proxyConfig.Node = "sidecar~1.2.3.4~foo~bar"
			proxyConfig.PlatEnv = &fakePlatform{
				meta: c.platformMeta,
			}
			proxyConfig.LocalEnv = localEnv
			proxyConfig.InstanceIPs = []string{"10.3.3.3", "10.4.4.4", "10.5.5.5", "10.6.6.6", "10.4.4.4"}

			fn, err := New(BootStrapConfig{
				Metadata: proxyConfig,
			}).CreateFileForEpoch(0)
			if err != nil {
				t.Fatal(err)
			}

			read, err := ioutil.ReadFile(fn)
			if err != nil {
				t.Error("Error reading generated file ", err)
				return
			}
			fmt.Printf(string(read))
		})
	}
}

func loadProxyConfig(base, out string, _ *testing.T) (*model.NodeMetadata, error) {
	content, err := ioutil.ReadFile("testdata/" + base + ".json")
	if err != nil {
		return nil, err
	}
	cfg := &model.NodeMetadata{}
	err = json.Unmarshal(content, cfg)
	if err != nil {
		return nil, err
	}

	cfg.CustomConfigFile = "testdata/grpc_bootstrap.json"

	// Exported from makefile or env
	cfg.ConfigPath = out + "/bootstrap/" + base
	return cfg, nil
}

// createEnv takes labels and annotations are returns environment in go format.
func createEnv(t *testing.T, labels map[string]string, anno map[string]string) (map[string]string, []string) {
	merged := map[string]string{}
	mergeMap(merged, labels)
	mergeMap(merged, anno)

	envs := make([]string, 0)

	if labels != nil {
		envs = append(envs, encodeAsJSON(t, labels, "LABELS"))
	}

	if anno != nil {
		envs = append(envs, encodeAsJSON(t, anno, "ANNOTATIONS"))
	}
	return merged, envs
}

func encodeAsJSON(t *testing.T, data map[string]string, name string) string {
	jsonStr, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("failed to marshal %s %v: %v", name, data, err)
	}
	return name + "=" + string(jsonStr)
}

func TestIsIPv6Proxy(t *testing.T) {
	tests := []struct {
		name     string
		addrs    []string
		expected bool
	}{
		{
			name:     "ipv4 only",
			addrs:    []string{"1.1.1.1", "127.0.0.1", "2.2.2.2"},
			expected: false,
		},
		{
			name:     "ipv6 only",
			addrs:    []string{"1111:2222::1", "::1", "2222:3333::1"},
			expected: true,
		},
		{
			name:     "mixed ipv4 and ipv6",
			addrs:    []string{"1111:2222::1", "::1", "127.0.0.1", "2.2.2.2", "2222:3333::1"},
			expected: false,
		},
	}
	for _, tt := range tests {
		result := isIPv6Proxy(tt.addrs)
		if result != tt.expected {
			t.Errorf("Test %s failed, expected: %t got: %t", tt.name, tt.expected, result)
		}
	}
}

func mergeMap(to map[string]string, from map[string]string) {
	for k, v := range from {
		to[k] = v
	}
}

type fakePlatform struct {
	platform.Environment

	meta   map[string]string
	labels map[string]string
}

func (f *fakePlatform) Metadata() map[string]string {
	return f.meta
}

func (f *fakePlatform) Locality() *core.Locality {
	return &core.Locality{}
}

func (f *fakePlatform) Labels() map[string]string {
	return f.labels
}

func (f *fakePlatform) IsKubernetes() bool {
	return true
}
