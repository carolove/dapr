// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package runtime

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	global_config "github.com/dapr/dapr/pkg/config"
	"github.com/dapr/dapr/pkg/grpc"
	"github.com/dapr/dapr/pkg/logger"
	"github.com/dapr/dapr/pkg/metrics"
	"github.com/dapr/dapr/pkg/modes"
	"github.com/dapr/dapr/pkg/operator/client"
	"github.com/dapr/dapr/pkg/runtime/security"
	"github.com/dapr/dapr/pkg/version"
	"github.com/pkg/errors"
)

// FromFlags parses command flags and returns DaprRuntime instance
func FromFlags(mode, daprHTTPPort,daprAPIGRPCPort,daprInternalGRPCPort,appPort,profilePort,appProtocol,componentsPath,config,appID,controlPlaneAddress,sentryAddress,
	placementServiceHostAddr,allowedOrigins string, enableProfiling,runtimeVersion bool,appMaxConcurrency int,enableMTLS,appSSL bool, flagLogLevel string, logAsJson bool, metricsPort string) (*DaprRuntime, error) {

	loggerOptions := logger.DefaultOptions()
	metricsExporter := metrics.NewExporter(metrics.DefaultMetricNamespace)


	loggerOptions.OutputLevel = flagLogLevel
	loggerOptions.JSONFormatEnabled = logAsJson
	metricsExporter.Options().MetricsPort = metricsPort

	if runtimeVersion {
		fmt.Println(version.Version())
		os.Exit(0)
	}

	if appID == "" {
		return nil, errors.New("app-id parameter cannot be empty")
	}

	// Apply options to all loggers
	loggerOptions.SetAppID(appID)
	if err := logger.ApplyOptionsToLoggers(&loggerOptions); err != nil {
		return nil, err
	}

	log.Infof("starting Dapr Runtime -- version %s -- commit %s", version.Version(), version.Commit())
	log.Infof("log level set to: %s", loggerOptions.OutputLevel)

	// Initialize dapr metrics exporter
	if err := metricsExporter.Init(); err != nil {
		log.Fatal(err)
	}

	daprHTTP, err := strconv.Atoi(daprHTTPPort)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing dapr-http-port flag")
	}

	daprAPIGRPC, err := strconv.Atoi(daprAPIGRPCPort)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing dapr-grpc-port flag")
	}

	profPort, err := strconv.Atoi(profilePort)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing profile-port flag")
	}

	var daprInternalGRPC int
	if daprInternalGRPCPort != "" {
		daprInternalGRPC, err = strconv.Atoi(daprInternalGRPCPort)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing dapr-internal-grpc-port")
		}
	} else {
		daprInternalGRPC, err = grpc.GetFreePort()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get free port for internal grpc server")
		}
	}

	var applicationPort int
	if appPort != "" {
		applicationPort, err = strconv.Atoi(appPort)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing app-port")
		}
	}

	placementAddresses := []string{}
	if placementServiceHostAddr != "" {
		placementAddresses = parsePlacementAddr(placementServiceHostAddr)
	}

	var concurrency int
	if appMaxConcurrency != -1 {
		concurrency = appMaxConcurrency
	}

	appPrtcl := string(HTTPProtocol)
	if appProtocol != string(HTTPProtocol) {
		appPrtcl = appProtocol
	}

	runtimeConfig := NewRuntimeConfig(appID, placementAddresses, controlPlaneAddress, allowedOrigins, config, componentsPath,
		appPrtcl, mode, daprHTTP, daprInternalGRPC, daprAPIGRPC, applicationPort, profPort, enableProfiling, concurrency, enableMTLS, sentryAddress, appSSL)

	var globalConfig *global_config.Configuration
	var configErr error

	if enableMTLS {
		runtimeConfig.CertChain, err = security.GetCertChain()
		if err != nil {
			return nil, err
		}
	}

	var accessControlList *global_config.AccessControlList
	var namespace string

	if config != "" {
		switch modes.DaprMode(mode) {
		case modes.KubernetesMode:
			client, conn, clientErr := client.GetOperatorClient(controlPlaneAddress, security.TLSServerName, runtimeConfig.CertChain)
			if clientErr != nil {
				return nil, clientErr
			}
			defer conn.Close()
			namespace = os.Getenv("NAMESPACE")
			globalConfig, configErr = global_config.LoadKubernetesConfiguration(config, namespace, client)
		case modes.StandaloneMode:
			globalConfig, _, configErr = global_config.LoadStandaloneConfiguration(config)
		}

		if configErr != nil {
			log.Debugf("Config error: %v", configErr)
		}
	}

	if configErr != nil {
		log.Fatalf("error loading configuration: %s", configErr)
	}
	if globalConfig == nil {
		log.Info("loading default configuration")
		globalConfig = global_config.LoadDefaultConfiguration()
	}

	accessControlList, err = global_config.ParseAccessControlSpec(globalConfig.Spec.AccessControlSpec)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return NewDaprRuntime(runtimeConfig, globalConfig, accessControlList), nil
}

func parsePlacementAddr(val string) []string {
	parsed := []string{}
	p := strings.Split(val, ",")
	for _, addr := range p {
		parsed = append(parsed, strings.TrimSpace(addr))
	}
	return parsed
}
