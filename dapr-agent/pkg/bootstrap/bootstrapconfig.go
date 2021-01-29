package bootstrap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/dapr/dapr/dapr-agent/pkg/model"
	"github.com/dapr/dapr/dapr-agent/pkg/option"
	"github.com/dapr/dapr/dapr-agent/pkg/platform"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"istio.io/istio/pkg/config/constants"
)

type BootStrapConfig struct {
	Metadata *model.NodeMetadata
}

// newTemplateParams creates a new template configuration for the given configuration.
func (cfg BootStrapConfig) toTemplateParams() (map[string]interface{}, error) {
	opts := make([]option.Instance, 0)

	// Fill in default config values.
	if cfg.Metadata.PlatEnv == nil {
		cfg.Metadata.PlatEnv = platform.Discover()
	}

	// Remove duplicates from the node IPs.
	cfg.Metadata.InstanceIPs = removeDuplicates(cfg.Metadata.InstanceIPs)

	opts = append(opts,
		option.NodeID(cfg.Metadata.Node),
		option.Cluster(cfg.Metadata.ServiceCluster),
		option.DiscoveryAddress(cfg.Metadata.DiscoveryAddress),
		option.DiscoveryChannelCreds("insecure"),
		option.EnabledFeatures([]string{"xds_v3"}),
		option.DiscoveryHost(cfg.Metadata.DiscoveryAddress))

	// Support passing extra info from node environment as metadata
	meta, rawMeta, err := getNodeMetaData(cfg.Metadata.LocalEnv, cfg.Metadata.PlatEnv, cfg.Metadata.InstanceIPs)
	if err != nil {
		return nil, err
	}
	opts = append(opts, getNodeMetadataOptions(meta, rawMeta, cfg.Metadata.PlatEnv)...)

	opts = append(opts,
		option.Localhost(option.LocalhostIPv4),
		option.Wildcard(option.WildcardIPv4),
		option.DNSLookupFamily(option.DNSLookupFamilyIPv4))

	// TODO: allow reading a file with additional metadata (for example if created with
	// 'envref'. This will allow dapr runtime to generate the right config even if the pod info
	// is not available (in particular in some multi-cluster cases)
	return option.NewTemplateParams(opts...)
}

func getLocalityOptions(meta *model.NodeMetadata, platEnv platform.Environment) []option.Instance {
	var l *core.Locality
	if meta.Labels[model.LocalityLabel] == "" {
		l = platEnv.Locality()
		// The locality string was not set, try to get locality from platform
	} else {
		localityString := model.GetLocalityLabelOrDefault(meta.Labels[model.LocalityLabel], "")
		l = model.ConvertLocality(localityString)
	}

	return []option.Instance{option.Region(l.Region), option.Zone(l.Zone), option.SubZone(l.SubZone)}
}

func getNodeMetadataOptions(meta *model.NodeMetadata, rawMeta map[string]interface{},
	platEnv platform.Environment) []option.Instance {
	// Add locality options.
	opts := getLocalityOptions(meta, platEnv)

	opts = append(opts, option.NodeMetadata(meta, rawMeta))
	return opts
}

type setMetaFunc func(m map[string]interface{}, key string, val string)

func extractMetadata(envs []string, prefix string, set setMetaFunc, meta map[string]interface{}) {
	metaPrefixLen := len(prefix)
	for _, e := range envs {
		if !shouldExtract(e, prefix) {
			continue
		}
		v := e[metaPrefixLen:]
		if !isEnvVar(v) {
			continue
		}
		metaKey, metaVal := parseEnvVar(v)
		set(meta, metaKey, metaVal)
	}
}

func shouldExtract(envVar, prefix string) bool {
	return strings.HasPrefix(envVar, prefix)
}

func isEnvVar(str string) bool {
	return strings.Contains(str, "=")
}

func parseEnvVar(varStr string) (string, string) {
	parts := strings.SplitN(varStr, "=", 2)
	if len(parts) != 2 {
		return varStr, ""
	}
	return parts[0], parts[1]
}

func jsonStringToMap(jsonStr string) (m map[string]string) {
	err := json.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		log.Warnf("Env variable with value %q failed json unmarshal: %v", jsonStr, err)
	}
	return
}

func extractAttributesMetadata(envVars []string, plat platform.Environment, meta *model.NodeMetadata) {
	for _, varStr := range envVars {
		name, val := parseEnvVar(varStr)
		switch name {
		case "POD_NAME":
			meta.InstanceName = val
		case "POD_NAMESPACE":
			meta.Namespace = val
		case "SERVICE_ACCOUNT":
			meta.ServiceAccount = val
		}
	}
	if plat != nil && len(plat.Metadata()) > 0 {
		meta.PlatformMetadata = plat.Metadata()
	}
}

// getNodeMetaData function uses an environment variable contract
func getNodeMetaData(envs []string, plat platform.Environment, nodeIPs []string) (*model.NodeMetadata, map[string]interface{}, error) {
	meta := &model.NodeMetadata{}
	untypedMeta := map[string]interface{}{}

	extractAttributesMetadata(envs, plat, meta)

	// Support multiple network interfaces, removing duplicates.
	meta.InstanceIPs = nodeIPs

	// Add all instance labels with lower precedence than pod labels
	extractInstanceLabels(plat, meta)

	// Add all pod labels found from filesystem
	// These are typically volume mounted by the downward API
	lbls, err := readPodLabels()
	if err == nil {
		if meta.Labels == nil {
			meta.Labels = map[string]string{}
		}
		for k, v := range lbls {
			meta.Labels[k] = v
		}
	} else {
		if os.IsNotExist(err) {
			log.Debugf("failed to read pod labels: %v", err)
		} else {
			log.Warnf("failed to read pod labels: %v", err)
		}
	}

	return meta, untypedMeta, nil
}

// Extracts instance labels for the platform into model.NodeMetadata.Labels
// only if not running on Kubernetes
func extractInstanceLabels(plat platform.Environment, meta *model.NodeMetadata) {
	if plat == nil || meta == nil || plat.IsKubernetes() {
		return
	}
	instanceLabels := plat.Labels()
	if meta.Labels == nil {
		meta.Labels = map[string]string{}
	}
	for k, v := range instanceLabels {
		meta.Labels[k] = v
	}
}

func readPodLabels() (map[string]string, error) {
	b, err := ioutil.ReadFile(constants.PodInfoLabelsPath)
	if err != nil {
		return nil, err
	}
	return ParseDownwardAPI(string(b))
}

// Fields are stored as format `%s=%q`, we will parse this back to a map
func ParseDownwardAPI(i string) (map[string]string, error) {
	res := map[string]string{}
	for _, line := range strings.Split(i, "\n") {
		sl := strings.SplitN(line, "=", 2)
		if len(sl) != 2 {
			continue
		}
		key := sl[0]
		// Strip the leading/trailing quotes

		val, err := strconv.Unquote(sl[1])
		if err != nil {
			return nil, fmt.Errorf("failed to unquote %v: %v", sl[1], err)
		}
		res[key] = val
	}
	return res, nil
}
func removeDuplicates(values []string) []string {
	set := make(map[string]struct{})
	newValues := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := set[v]; !ok {
			set[v] = struct{}{}
			newValues = append(newValues, v)
		}
	}
	return newValues
}
