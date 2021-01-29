package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	daprproxy "github.com/dapr/dapr/dapr-agent/pkg/agent"
	"github.com/dapr/dapr/dapr-agent/pkg/config"
	"github.com/dapr/dapr/dapr-agent/pkg/model"
	"github.com/dapr/dapr/pkg/cors"
	"github.com/dapr/dapr/pkg/modes"
	"github.com/dapr/dapr/pkg/runtime"
	"github.com/dapr/dapr/pkg/version"
	"istio.io/istio/pilot/pkg/util/network"
	"istio.io/istio/pkg/cmd"
	"istio.io/istio/pkg/config/constants"
	"istio.io/pkg/env"

	"github.com/spf13/cobra"
)

const (
	localHostIPv4 = "127.0.0.1"
	localHostIPv6 = "[::1]"
)

// TODO: Move most of this to pkg.

var (
	role model.NodeMetadata

	instanceIPVar     = env.RegisterStringVar("INSTANCE_IP", "", "")
	podNameVar        = env.RegisterStringVar("POD_NAME", "", "")
	podNamespaceVar   = env.RegisterStringVar("POD_NAMESPACE", "", "")
	serviceAccountVar = env.RegisterStringVar("SERVICE_ACCOUNT", "", "Name of service account")

	proxyConfigEnv = env.RegisterStringVar(
		"PROXY_CONFIG",
		"",
		"The proxy configuration. This will be set by the injection - gateways will use file mounts.",
	).Get()

	rootCmd = &cobra.Command{
		Use:          "dapr-agent",
		Short:        "dapr agent.",
		Long:         "dapr agent runs in the sidecar or gateway container and bootstraps dapr-runtime.",
		SilenceUsage: true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
	}

	proxyCmd = &cobra.Command{
		Use:   "proxy",
		Short: "dapr runtime agent",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			// Allow unknown flags for backward-compatibility.
			UnknownFlags: true,
		},
		PersistentPreRunE: configureLogging,
		RunE: func(c *cobra.Command, args []string) error {

			// Extract pod variables.
			role.PodName = podNameVar.Get()
			role.PodNamespace = podNamespaceVar.Get()
			podIP := net.ParseIP(instanceIPVar.Get()) // protobuf encoding of IP_ADDRESS e

			log.Infof("Version %s", version.Version())

			proxyConfig, err := config.ConstructProxyConfig(&role)
			if err != nil {
				return fmt.Errorf("failed to get proxy config: %v", err)
			}

			if podIP != nil {
				proxyConfig.InstanceIPs = []string{podIP.String()}
			}

			// Obtain all the IPs from the node
			if ipAddrs, ok := network.GetPrivateIPs(context.Background()); ok {
				if len(proxyConfig.InstanceIPs) == 1 {
					for _, ip := range ipAddrs {
						// prevent duplicate ips, the first one must be the pod ip
						// as we pick the first ip as pod ip in dapr runtime
						if proxyConfig.InstanceIPs[0] != ip {
							proxyConfig.InstanceIPs = append(proxyConfig.InstanceIPs, ip)
						}
					}
				} else {
					proxyConfig.InstanceIPs = append(proxyConfig.InstanceIPs, ipAddrs...)
				}
			}

			// No IP addresses provided, append 127.0.0.1 for ipv4 and ::1 for ipv6
			if len(proxyConfig.InstanceIPs) == 0 {
				proxyConfig.InstanceIPs = append(proxyConfig.InstanceIPs, "127.0.0.1")
				proxyConfig.InstanceIPs = append(proxyConfig.InstanceIPs, "::1")
			}
			// node is related with node ips
			if proxyConfig.Node == "" {
				proxyConfig.Node = proxyConfig.ServiceNode()
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Legacy environment variable is set, us that instead
			drainDuration := time.Second * time.Duration(1)

			daprRuntime := daprproxy.NewDaprd(proxyConfig)
			agent := daprproxy.NewAgent(daprRuntime, drainDuration)

			// Watcher is also kicking dapr runtime start.
			watcher := daprproxy.NewWatcher(agent.Restart)
			go watcher.Run(ctx)

			// On SIGINT or SIGTERM, cancel the context, triggering a graceful shutdown
			go cmd.WaitSignalFunc(cancel)

			return agent.Run(ctx)
		},
	}
)

func configureLogging(_ *cobra.Command, _ []string) error {

	return nil
}

func init() {
	// Flags for proxy configuration
	proxyCmd.PersistentFlags().StringVar(&role.ServiceCluster, "serviceCluster", constants.ServiceClusterName, "Service cluster")
	proxyCmd.PersistentFlags().StringVar(&role.LogLevel, "proxyLogLevel", "warning",
		fmt.Sprintf("The log level used to start the dapr runtime (choose from {%s, %s, %s, %s, %s, %s, %s})",
			"trace", "debug", "info", "warning", "error", "critical", "off"))
	proxyCmd.PersistentFlags().StringVar(&role.ComponentLogLevel, "proxyComponentLogLevel", "misc:error",
		"The component log level used to start the dapr runtime")

	// dapr runtime
	proxyCmd.PersistentFlags().StringVar(&role.CustomConfigFile, "customBootstrapTmpl", "/var/lib/dapr/runtime/grpc_bootstrap_tmpl.json",
		"custom bootstrap config template")
	proxyCmd.PersistentFlags().StringVar(&role.BinaryPath, "customBinaryPath", "/usr/local/bin/daprd",
		"custom dapr runtime binary path")
	proxyCmd.PersistentFlags().StringVar(&role.Mode, "mode", string(modes.StandaloneMode), "Runtime mode for Dapr")
	proxyCmd.PersistentFlags().BoolVar(&role.EnableMesh, "enable-mesh", false, "Enable Service Mesh")
	proxyCmd.PersistentFlags().StringVar(&role.DaprHTTPPort, "dapr-http-port", fmt.Sprintf("%v", runtime.DefaultDaprHTTPPort),
		"HTTP port for Dapr API to listen on")
	proxyCmd.PersistentFlags().StringVar(&role.DaprAPIGRPCPort, "dapr-grpc-port", fmt.Sprintf("%v", runtime.DefaultDaprAPIGRPCPort),
		"gRPC port for the Dapr API to listen on")
	proxyCmd.PersistentFlags().StringVar(&role.DaprInternalGRPCPort, "dapr-internal-grpc-port", "",
		"gRPC port for the Dapr Internal API to listen on")
	proxyCmd.PersistentFlags().StringVar(&role.AppPort, "app-port", "", "The port the application is listening on")
	proxyCmd.PersistentFlags().StringVar(&role.ProfilePort, "profile-port", fmt.Sprintf("%v", runtime.DefaultProfilePort),
		"The port for the profile server")
	proxyCmd.PersistentFlags().StringVar(&role.AppProtocol, "app-protocol", string(runtime.HTTPProtocol),
		"Protocol for the application: grpc or http")
	proxyCmd.PersistentFlags().StringVar(&role.ComponentsPath, "components-path", "",
		"Path for components directory. If empty, components will not be loaded. Self-hosted mode only")
	//proxyCmd.PersistentFlags().StringVar(&config ,"config", "", "Path to config file, or name of a configuration object")
	proxyCmd.PersistentFlags().StringVar(&role.AppID, "app-id", "",
		"A unique ID for Dapr. Used for Service Discovery and state")
	proxyCmd.PersistentFlags().StringVar(&role.ControlPlaneAddress, "control-plane-address", "", "Address for a Dapr control plane")
	proxyCmd.PersistentFlags().StringVar(&role.DiscoveryAddress, "discovery-plane-address", "xds-server.dapr-system.svc.cluster.local:15010", "Address for a grpc xds discovery control plane")
	proxyCmd.PersistentFlags().StringVar(&role.SentryAddress, "sentry-address", "", "Address for the Sentry CA service")
	proxyCmd.PersistentFlags().StringVar(&role.PlacementServiceHostAddr, "placement-host-address", "", "Addresses for Dapr Actor Placement servers")
	proxyCmd.PersistentFlags().StringVar(&role.AllowedOrigins, "allowed-origins", cors.DefaultAllowedOrigins, "Allowed HTTP origins")
	proxyCmd.PersistentFlags().BoolVar(&role.EnableProfiling, "enable-profiling", false, "Enable profiling")
	proxyCmd.PersistentFlags().BoolVar(&role.RuntimeVersion, "version", false, "Prints the runtime version")
	proxyCmd.PersistentFlags().IntVar(&role.AppMaxConcurrency, "app-max-concurrency", -1,
		"Controls the concurrency level when forwarding requests to user code")
	proxyCmd.PersistentFlags().BoolVar(&role.EnableMTLS, "enable-mtls", false, "Enables automatic mTLS for daprd to daprd communication channels")
	proxyCmd.PersistentFlags().BoolVar(&role.AppSSL, "app-ssl", false, "Sets the URI scheme of the app to https and attempts an SSL connection")
	proxyCmd.PersistentFlags().BoolVar(&role.Sidecar, "sidecar-enabled", true, "Sets daprd proxy mode, default sidecar enabled")
	proxyCmd.PersistentFlags().IntVar(&role.DaprHTTPMaxRequestSize, "dapr-http-max-request-size", -1,
		"Increasing max size of request body in MB to handle uploading of big files. By default 4 MB.")

	cmd.AddFlags(rootCmd)

	rootCmd.AddCommand(proxyCmd)
}

// TODO: get the config and bootstrap from dapr runtime, by passing the env

// Use env variables - from injection, k8s and local namespace config map.
// No CLI parameters.
func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(-1)
	}
}

// isIPv6Proxy check the addresses slice and returns true for a valid IPv6 address
// for all other cases it returns false
func isIPv6Proxy(ipAddrs []string) bool {
	for i := 0; i < len(ipAddrs); i++ {
		addr := net.ParseIP(ipAddrs[i])
		if addr == nil {
			// Should not happen, invalid IP in proxy's IPAddresses slice should have been caught earlier,
			// skip it to prevent a panic.
			continue
		}
		if addr.To4() != nil {
			return false
		}
	}
	return true
}
