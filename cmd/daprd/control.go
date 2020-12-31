package main

import (
	"fmt"
	"strings"

	rawhttp "net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
	rawruntime "runtime"

	"github.com/dapr/dapr/pkg/cors"
	"github.com/dapr/dapr/pkg/metrics"
	"github.com/dapr/dapr/pkg/modes"
	"github.com/dapr/dapr/pkg/runtime"
	"github.com/urfave/cli"

	"mosn.io/mosn/pkg/admin/store"
	"mosn.io/mosn/pkg/configmanager"
	"mosn.io/mosn/pkg/featuregate"
	mosnlog "mosn.io/mosn/pkg/log"
	mosnmetrics "mosn.io/mosn/pkg/metrics"
	"mosn.io/mosn/pkg/mosn"
	"mosn.io/mosn/pkg/types"

	// Included components in compiled daprd

	// Secret stores
	"github.com/dapr/components-contrib/secretstores"
	"github.com/dapr/components-contrib/secretstores/aws/secretmanager"
	"github.com/dapr/components-contrib/secretstores/azure/keyvault"
	gcp_secretmanager "github.com/dapr/components-contrib/secretstores/gcp/secretmanager"
	"github.com/dapr/components-contrib/secretstores/hashicorp/vault"
	sercetstores_kubernetes "github.com/dapr/components-contrib/secretstores/kubernetes"
	secretstore_env "github.com/dapr/components-contrib/secretstores/local/env"
	secretstore_file "github.com/dapr/components-contrib/secretstores/local/file"
	secretstores_loader "github.com/dapr/dapr/pkg/components/secretstores"

	// State Stores
	"github.com/dapr/components-contrib/state"
	"github.com/dapr/components-contrib/state/aerospike"
	state_dynamodb "github.com/dapr/components-contrib/state/aws/dynamodb"
	state_azure_blobstorage "github.com/dapr/components-contrib/state/azure/blobstorage"
	state_cosmosdb "github.com/dapr/components-contrib/state/azure/cosmosdb"
	state_azure_tablestorage "github.com/dapr/components-contrib/state/azure/tablestorage"
	"github.com/dapr/components-contrib/state/cassandra"
	"github.com/dapr/components-contrib/state/cloudstate"
	"github.com/dapr/components-contrib/state/couchbase"
	"github.com/dapr/components-contrib/state/gcp/firestore"
	"github.com/dapr/components-contrib/state/hashicorp/consul"
	"github.com/dapr/components-contrib/state/hazelcast"
	"github.com/dapr/components-contrib/state/memcached"
	"github.com/dapr/components-contrib/state/mongodb"
	"github.com/dapr/components-contrib/state/postgresql"
	state_redis "github.com/dapr/components-contrib/state/redis"
	"github.com/dapr/components-contrib/state/rethinkdb"
	"github.com/dapr/components-contrib/state/sqlserver"
	"github.com/dapr/components-contrib/state/zookeeper"
	state_loader "github.com/dapr/dapr/pkg/components/state"

	// Pub/Sub
	pubs "github.com/dapr/components-contrib/pubsub"
	pubsub_snssqs "github.com/dapr/components-contrib/pubsub/aws/snssqs"
	pubsub_eventhubs "github.com/dapr/components-contrib/pubsub/azure/eventhubs"
	"github.com/dapr/components-contrib/pubsub/azure/servicebus"
	pubsub_gcp "github.com/dapr/components-contrib/pubsub/gcp/pubsub"
	pubsub_hazelcast "github.com/dapr/components-contrib/pubsub/hazelcast"
	pubsub_kafka "github.com/dapr/components-contrib/pubsub/kafka"
	pubsub_mqtt "github.com/dapr/components-contrib/pubsub/mqtt"
	"github.com/dapr/components-contrib/pubsub/natsstreaming"
	pubsub_pulsar "github.com/dapr/components-contrib/pubsub/pulsar"
	"github.com/dapr/components-contrib/pubsub/rabbitmq"
	pubsub_redis "github.com/dapr/components-contrib/pubsub/redis"
	pubsub_loader "github.com/dapr/dapr/pkg/components/pubsub"

	// Name resolutions
	nr "github.com/dapr/components-contrib/nameresolution"
	nr_kubernetes "github.com/dapr/components-contrib/nameresolution/kubernetes"
	nr_mdns "github.com/dapr/components-contrib/nameresolution/mdns"
	nr_loader "github.com/dapr/dapr/pkg/components/nameresolution"
	nr_mesh "github.com/dapr/dapr/pkg/components/nameresolution/mesh"

	// Bindings
	"github.com/dapr/components-contrib/bindings"
	"github.com/dapr/components-contrib/bindings/apns"
	"github.com/dapr/components-contrib/bindings/aws/dynamodb"
	"github.com/dapr/components-contrib/bindings/aws/kinesis"
	"github.com/dapr/components-contrib/bindings/aws/s3"
	"github.com/dapr/components-contrib/bindings/aws/sns"
	"github.com/dapr/components-contrib/bindings/aws/sqs"
	"github.com/dapr/components-contrib/bindings/azure/blobstorage"
	bindings_cosmosdb "github.com/dapr/components-contrib/bindings/azure/cosmosdb"
	"github.com/dapr/components-contrib/bindings/azure/eventgrid"
	"github.com/dapr/components-contrib/bindings/azure/eventhubs"
	"github.com/dapr/components-contrib/bindings/azure/servicebusqueues"
	"github.com/dapr/components-contrib/bindings/azure/signalr"
	"github.com/dapr/components-contrib/bindings/azure/storagequeues"
	"github.com/dapr/components-contrib/bindings/cron"
	"github.com/dapr/components-contrib/bindings/gcp/bucket"
	"github.com/dapr/components-contrib/bindings/gcp/pubsub"
	"github.com/dapr/components-contrib/bindings/http"
	"github.com/dapr/components-contrib/bindings/influx"
	"github.com/dapr/components-contrib/bindings/kafka"
	"github.com/dapr/components-contrib/bindings/kubernetes"
	"github.com/dapr/components-contrib/bindings/mqtt"
	"github.com/dapr/components-contrib/bindings/postgres"
	"github.com/dapr/components-contrib/bindings/postmark"
	bindings_rabbitmq "github.com/dapr/components-contrib/bindings/rabbitmq"
	"github.com/dapr/components-contrib/bindings/redis"
	"github.com/dapr/components-contrib/bindings/rethinkdb/statechange"
	"github.com/dapr/components-contrib/bindings/twilio/sendgrid"
	"github.com/dapr/components-contrib/bindings/twilio/sms"
	"github.com/dapr/components-contrib/bindings/twitter"
	bindings_loader "github.com/dapr/dapr/pkg/components/bindings"

	// HTTP Middleware
	middleware "github.com/dapr/components-contrib/middleware"
	"github.com/dapr/components-contrib/middleware/http/bearer"
	"github.com/dapr/components-contrib/middleware/http/oauth2"
	"github.com/dapr/components-contrib/middleware/http/oauth2clientcredentials"
	"github.com/dapr/components-contrib/middleware/http/opa"
	"github.com/dapr/components-contrib/middleware/http/ratelimit"
	http_middleware_loader "github.com/dapr/dapr/pkg/components/middleware/http"
	http_middleware "github.com/dapr/dapr/pkg/middleware/http"
	"github.com/valyala/fasthttp"
)

var (
	flagToMosnLogLevel = map[string]string{
		"trace":    "TRACE",
		"debug":    "DEBUG",
		"info":     "INFO",
		"warning":  "WARN",
		"error":    "ERROR",
		"critical": "FATAL",
		"off":      "OFF",
	}

	cmdStart = cli.Command{
		Name:  "start",
		Usage: "start mecha sidecar",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "mode",
				Usage: "Runtime mode for mecha",
				Value: string(modes.StandaloneMode),
			}, cli.StringFlag{
				Name:  "dapr-http-port",
				Usage: "HTTP port for Dapr API to listen on",
				Value: fmt.Sprintf("%v", runtime.DefaultDaprHTTPPort),
			}, cli.StringFlag{
				Name:  "dapr-grpc-port",
				Usage: "gRPC port for the Dapr API to listen on",
				Value: fmt.Sprintf("%v", runtime.DefaultDaprAPIGRPCPort),
			}, cli.StringFlag{
				Name:  "dapr-internal-grpc-port",
				Usage: "gRPC port for the Dapr Internal API to listen on",
				Value: "",
			}, cli.StringFlag{
				Name:  "app-port",
				Usage: "The port the application is listening on",
				Value: "",
			}, cli.StringFlag{
				Name:  "profile-port",
				Usage: "The port for the profile server",
				Value: fmt.Sprintf("%v", runtime.DefaultProfilePort),
			}, cli.StringFlag{
				Name:  "app-protocol",
				Usage: "Protocol for the application: grpc or http",
				Value: string(runtime.HTTPProtocol),
			}, cli.StringFlag{
				Name:  "components-path",
				Usage: "Path for components directory. If empty, components will not be loaded. Self-hosted mode only",
				Value: "",
			}, cli.StringFlag{
				Name:  "config",
				Usage: "Path to config file, or name of a configuration object",
				Value: "",
			}, cli.StringFlag{
				Name:  "app-id",
				Usage: "A unique ID for mecha. Used for Service Discovery and state",
				Value: "",
			}, cli.StringFlag{
				Name:  "control-plane-address",
				Usage: "Address for a mecha control plane",
				Value: "",
			}, cli.StringFlag{
				Name:  "sentry-address",
				Usage: "Address for the Sentry CA service",
				Value: "",
			}, cli.StringFlag{
				Name:  "placement-host-address",
				Usage: "Addresses for mecha Actor Placement servers",
				Value: "",
			}, cli.StringFlag{
				Name:  "allowed-origins",
				Usage: "Allowed HTTP origins",
				Value: cors.DefaultAllowedOrigins,
			}, cli.BoolFlag{
				Name:  "enable-profiling",
				Usage: "Enable profiling",
			}, cli.IntFlag{
				Name:  "app-max-concurrency",
				Usage: "Controls the concurrency level when forwarding requests to user code",
				Value: -1,
			}, cli.BoolFlag{
				Name:  "enable-mtls",
				Usage: "Enables automatic mTLS for mechad to mechad communication channels",
			}, cli.BoolFlag{
				Name:  "app-ssl",
				Usage: "Sets the URI scheme of the app to https and attempts an SSL connection",
			}, cli.BoolFlag{
				Name:  "log-as-json",
				Usage: "print log as JSON (default false)",
			}, cli.StringFlag{
				Name:  "metrics-port",
				Usage: "The port for the metrics server",
				Value: metrics.DefaultMetricsPort,
			},

			cli.StringFlag{
				Name:   "mosn-config, mc",
				Usage:  "Load configuration from `FILE`",
				EnvVar: "MOSN_CONFIG",
				Value:  "configs/mosn_config.json",
			}, cli.StringFlag{
				Name:   "service-cluster, s",
				Usage:  "sidecar service cluster",
				EnvVar: "SERVICE_CLUSTER",
			}, cli.StringFlag{
				Name:   "service-node, n",
				Usage:  "sidecar service node",
				EnvVar: "SERVICE_NODE",
			}, cli.StringFlag{
				Name:   "service-type, p",
				Usage:  "sidecar service type",
				EnvVar: "SERVICE_TYPE",
			}, cli.StringSliceFlag{
				Name:   "service-meta, sm",
				Usage:  "sidecar service metadata",
				EnvVar: "SERVICE_META",
			}, cli.StringSliceFlag{
				Name:   "service-lables, sl",
				Usage:  "sidecar service metadata labels",
				EnvVar: "SERVICE_LAB",
			}, cli.StringSliceFlag{
				Name:   "cluster-domain, domain",
				Usage:  "sidecar service metadata labels",
				EnvVar: "CLUSTER_DOMAIN",
			}, cli.StringFlag{
				Name:   "feature-gates, f",
				Usage:  "config feature gates",
				EnvVar: "FEATURE_GATES",
			}, cli.StringFlag{
				Name:   "pod-namespace, pns",
				Usage:  "mecha pod namespaces",
				EnvVar: "POD_NAMESPACE",
			}, cli.StringFlag{
				Name:   "pod-name, pn",
				Usage:  "mecha pod name",
				EnvVar: "POD_NAME",
			}, cli.StringFlag{
				Name:   "pod-ip, pi",
				Usage:  "mecha pod ip",
				EnvVar: "POD_IP",
			}, cli.StringFlag{
				Name:   "log-level, l",
				Usage:  "mecha log level, trace|debug|info|warning|error|critical|off",
				EnvVar: "LOG_LEVEL",
			}, cli.StringFlag{
				Name:  "log-format, lf",
				Usage: "mecha log format, currently useless",
			}, cli.StringSliceFlag{
				Name:  "component-log-level, lc",
				Usage: "mecha component format, currently useless",
			}, cli.StringFlag{
				Name:  "local-address-ip-version",
				Usage: "ip version, v4 or v6, currently useless",
			}, cli.IntFlag{
				Name:  "restart-epoch",
				Usage: "eporch to restart, align to Istio startup params, currently useless",
			}, cli.IntFlag{
				Name:  "drain-time-s",
				Usage: "seconds to drain, align to Istio startup params, currently useless",
			}, cli.StringFlag{
				Name:  "parent-shutdown-time-s",
				Usage: "parent shutdown time seconds, align to Istio startup params, currently useless",
			}, cli.IntFlag{
				Name:  "max-obj-name-len",
				Usage: "object name limit, align to Istio startup params, currently useless",
			}, cli.IntFlag{
				Name:  "concurrency",
				Usage: "concurrency, align to Istio startup params, currently useless",
			},
		},
		Action: func(c *cli.Context) error {
			// mecha flags
			mode := c.String("mode")
			daprHTTPPort := c.String("dapr-http-port")
			daprAPIGRPCPort := c.String("dapr-grpc-port")
			daprInternalGRPCPort := c.String("dapr-internal-grpc-port")
			appPort := c.String("app-port")
			profilePort := c.String("profile-port")
			appProtocol := c.String("app-protocol")
			componentsPath := c.String("components-path")
			config := c.String("config")
			appID := c.String("app-id")
			controlPlaneAddress := c.String("control-plane-address")
			sentryAddress := c.String("sentry-address")
			placementServiceHostAddr := c.String("placement-host-address")
			allowedOrigins := c.String("allowed-origins")
			enableProfiling := c.Bool("enable-profiling")
			runtimeVersion := c.Bool("version")
			appMaxConcurrency := c.Int("app-max-concurrency")
			enableMTLS := c.Bool("enable-mtls")
			appSSL := c.Bool("app-ssl")
			logAsJson := c.Bool("log-as-json")
			metricsPort := c.String("metrics-port")

			// mosn flags
			configPath := c.String("mosn-config")
			serviceCluster := c.String("service-cluster")
			serviceNode := c.String("service-node")
			serviceType := c.String("service-type")
			serviceMeta := c.StringSlice("service-meta")
			serviceLabels := c.StringSlice("service-lables")
			clusterDomain := c.String("cluster-domain")
			podName := c.String("pod-name")
			podNamespace := c.String("pod-namespace")
			podIp := c.String("pod-ip")
			flagLogLevel := c.String("log-level")
			featureGates := c.String("feature-gates")

			rt, err := runtime.FromFlags(mode, daprHTTPPort, daprAPIGRPCPort, daprInternalGRPCPort, appPort, profilePort, appProtocol, componentsPath, config,
				appID, controlPlaneAddress, sentryAddress, placementServiceHostAddr, allowedOrigins, enableProfiling,
				runtimeVersion, appMaxConcurrency, enableMTLS, appSSL, flagLogLevel, logAsJson, metricsPort)

			if err != nil {
				log.Fatal(err)
			}

			err = rt.Run(
				runtime.WithSecretStores(
					secretstores_loader.New("kubernetes", func() secretstores.SecretStore {
						return sercetstores_kubernetes.NewKubernetesSecretStore(logContrib)
					}),
					secretstores_loader.New("azure.keyvault", func() secretstores.SecretStore {
						return keyvault.NewAzureKeyvaultSecretStore(logContrib)
					}),
					secretstores_loader.New("hashicorp.vault", func() secretstores.SecretStore {
						return vault.NewHashiCorpVaultSecretStore(logContrib)
					}),
					secretstores_loader.New("aws.secretmanager", func() secretstores.SecretStore {
						return secretmanager.NewSecretManager(logContrib)
					}),
					secretstores_loader.New("gcp.secretmanager", func() secretstores.SecretStore {
						return gcp_secretmanager.NewSecreteManager(logContrib)
					}),
					secretstores_loader.New("local.file", func() secretstores.SecretStore {
						return secretstore_file.NewLocalSecretStore(logContrib)
					}),
					secretstores_loader.New("local.env", func() secretstores.SecretStore {
						return secretstore_env.NewEnvSecretStore(logContrib)
					}),
				),
				runtime.WithStates(
					state_loader.New("redis", func() state.Store {
						return state_redis.NewRedisStateStore(logContrib)
					}),
					state_loader.New("consul", func() state.Store {
						return consul.NewConsulStateStore(logContrib)
					}),
					state_loader.New("azure.blobstorage", func() state.Store {
						return state_azure_blobstorage.NewAzureBlobStorageStore(logContrib)
					}),
					state_loader.New("azure.cosmosdb", func() state.Store {
						return state_cosmosdb.NewCosmosDBStateStore(logContrib)
					}),
					state_loader.New("azure.tablestorage", func() state.Store {
						return state_azure_tablestorage.NewAzureTablesStateStore(logContrib)
					}),
					state_loader.New("cassandra", func() state.Store {
						return cassandra.NewCassandraStateStore(logContrib)
					}),
					state_loader.New("memcached", func() state.Store {
						return memcached.NewMemCacheStateStore(logContrib)
					}),
					state_loader.New("mongodb", func() state.Store {
						return mongodb.NewMongoDB(logContrib)
					}),
					state_loader.New("zookeeper", func() state.Store {
						return zookeeper.NewZookeeperStateStore(logContrib)
					}),
					state_loader.New("gcp.firestore", func() state.Store {
						return firestore.NewFirestoreStateStore(logContrib)
					}),
					state_loader.New("postgresql", func() state.Store {
						return postgresql.NewPostgreSQLStateStore(logContrib)
					}),
					state_loader.New("sqlserver", func() state.Store {
						return sqlserver.NewSQLServerStateStore(logContrib)
					}),
					state_loader.New("hazelcast", func() state.Store {
						return hazelcast.NewHazelcastStore(logContrib)
					}),
					state_loader.New("cloudstate.crdt", func() state.Store {
						return cloudstate.NewCRDT(logContrib)
					}),
					state_loader.New("couchbase", func() state.Store {
						return couchbase.NewCouchbaseStateStore(logContrib)
					}),
					state_loader.New("aerospike", func() state.Store {
						return aerospike.NewAerospikeStateStore(logContrib)
					}),
					state_loader.New("rethinkdb", func() state.Store {
						return rethinkdb.NewRethinkDBStateStore(logContrib)
					}),
					state_loader.New("aws.dynamodb", state_dynamodb.NewDynamoDBStateStore),
				),
				runtime.WithPubSubs(
					pubsub_loader.New("redis", func() pubs.PubSub {
						return pubsub_redis.NewRedisStreams(logContrib)
					}),
					pubsub_loader.New("natsstreaming", func() pubs.PubSub {
						return natsstreaming.NewNATSStreamingPubSub(logContrib)
					}),
					pubsub_loader.New("azure.eventhubs", func() pubs.PubSub {
						return pubsub_eventhubs.NewAzureEventHubs(logContrib)
					}),
					pubsub_loader.New("azure.servicebus", func() pubs.PubSub {
						return servicebus.NewAzureServiceBus(logContrib)
					}),
					pubsub_loader.New("rabbitmq", func() pubs.PubSub {
						return rabbitmq.NewRabbitMQ(logContrib)
					}),
					pubsub_loader.New("hazelcast", func() pubs.PubSub {
						return pubsub_hazelcast.NewHazelcastPubSub(logContrib)
					}),
					pubsub_loader.New("gcp.pubsub", func() pubs.PubSub {
						return pubsub_gcp.NewGCPPubSub(logContrib)
					}),
					pubsub_loader.New("kafka", func() pubs.PubSub {
						return pubsub_kafka.NewKafka(logContrib)
					}),
					pubsub_loader.New("snssqs", func() pubs.PubSub {
						return pubsub_snssqs.NewSnsSqs(logContrib)
					}),
					pubsub_loader.New("mqtt", func() pubs.PubSub {
						return pubsub_mqtt.NewMQTTPubSub(logContrib)
					}),
					pubsub_loader.New("pulsar", func() pubs.PubSub {
						return pubsub_pulsar.NewPulsar(logContrib)
					}),
				),
				runtime.WithNameResolutions(
					nr_loader.New("mdns", func() nr.Resolver {
						return nr_mdns.NewResolver(logContrib)
					}),
					nr_loader.New("kubernetes", func() nr.Resolver {
						return nr_kubernetes.NewResolver(logContrib)
					}),
					nr_loader.New("mesh", func() nr.Resolver {
						return nr_mesh.NewResolver(logContrib)
					}),
				),
				runtime.WithInputBindings(
					bindings_loader.NewInput("aws.sqs", func() bindings.InputBinding {
						return sqs.NewAWSSQS(logContrib)
					}),
					bindings_loader.NewInput("aws.kinesis", func() bindings.InputBinding {
						return kinesis.NewAWSKinesis(logContrib)
					}),
					bindings_loader.NewInput("azure.eventhubs", func() bindings.InputBinding {
						return eventhubs.NewAzureEventHubs(logContrib)
					}),
					bindings_loader.NewInput("kafka", func() bindings.InputBinding {
						return kafka.NewKafka(logContrib)
					}),
					bindings_loader.NewInput("mqtt", func() bindings.InputBinding {
						return mqtt.NewMQTT(logContrib)
					}),
					bindings_loader.NewInput("rabbitmq", func() bindings.InputBinding {
						return bindings_rabbitmq.NewRabbitMQ(logContrib)
					}),
					bindings_loader.NewInput("azure.servicebusqueues", func() bindings.InputBinding {
						return servicebusqueues.NewAzureServiceBusQueues(logContrib)
					}),
					bindings_loader.NewInput("azure.storagequeues", func() bindings.InputBinding {
						return storagequeues.NewAzureStorageQueues(logContrib)
					}),
					bindings_loader.NewInput("gcp.pubsub", func() bindings.InputBinding {
						return pubsub.NewGCPPubSub(logContrib)
					}),
					bindings_loader.NewInput("kubernetes", func() bindings.InputBinding {
						return kubernetes.NewKubernetes(logContrib)
					}),
					bindings_loader.NewInput("azure.eventgrid", func() bindings.InputBinding {
						return eventgrid.NewAzureEventGrid(logContrib)
					}),
					bindings_loader.NewInput("twitter", func() bindings.InputBinding {
						return twitter.NewTwitter(logContrib)
					}),
					bindings_loader.NewInput("cron", func() bindings.InputBinding {
						return cron.NewCron(logContrib)
					}),
					bindings_loader.NewInput("rethinkdb.statechange", func() bindings.InputBinding {
						return statechange.NewRethinkDBStateChangeBinding(logContrib)
					}),
				),
				runtime.WithOutputBindings(
					bindings_loader.NewOutput("apns", func() bindings.OutputBinding {
						return apns.NewAPNS(logContrib)
					}),
					bindings_loader.NewOutput("aws.sqs", func() bindings.OutputBinding {
						return sqs.NewAWSSQS(logContrib)
					}),
					bindings_loader.NewOutput("aws.sns", func() bindings.OutputBinding {
						return sns.NewAWSSNS(logContrib)
					}),
					bindings_loader.NewOutput("aws.kinesis", func() bindings.OutputBinding {
						return kinesis.NewAWSKinesis(logContrib)
					}),
					bindings_loader.NewOutput("azure.eventhubs", func() bindings.OutputBinding {
						return eventhubs.NewAzureEventHubs(logContrib)
					}),
					bindings_loader.NewOutput("aws.dynamodb", func() bindings.OutputBinding {
						return dynamodb.NewDynamoDB(logContrib)
					}),
					bindings_loader.NewOutput("azure.cosmosdb", func() bindings.OutputBinding {
						return bindings_cosmosdb.NewCosmosDB(logContrib)
					}),
					bindings_loader.NewOutput("gcp.bucket", func() bindings.OutputBinding {
						return bucket.NewGCPStorage(logContrib)
					}),
					bindings_loader.NewOutput("http", func() bindings.OutputBinding {
						return http.NewHTTP(logContrib)
					}),
					bindings_loader.NewOutput("kafka", func() bindings.OutputBinding {
						return kafka.NewKafka(logContrib)
					}),
					bindings_loader.NewOutput("mqtt", func() bindings.OutputBinding {
						return mqtt.NewMQTT(logContrib)
					}),
					bindings_loader.NewOutput("rabbitmq", func() bindings.OutputBinding {
						return bindings_rabbitmq.NewRabbitMQ(logContrib)
					}),
					bindings_loader.NewOutput("redis", func() bindings.OutputBinding {
						return redis.NewRedis(logContrib)
					}),
					bindings_loader.NewOutput("aws.s3", func() bindings.OutputBinding {
						return s3.NewAWSS3(logContrib)
					}),
					bindings_loader.NewOutput("azure.blobstorage", func() bindings.OutputBinding {
						return blobstorage.NewAzureBlobStorage(logContrib)
					}),
					bindings_loader.NewOutput("azure.servicebusqueues", func() bindings.OutputBinding {
						return servicebusqueues.NewAzureServiceBusQueues(logContrib)
					}),
					bindings_loader.NewOutput("azure.storagequeues", func() bindings.OutputBinding {
						return storagequeues.NewAzureStorageQueues(logContrib)
					}),
					bindings_loader.NewOutput("gcp.pubsub", func() bindings.OutputBinding {
						return pubsub.NewGCPPubSub(logContrib)
					}),
					bindings_loader.NewOutput("azure.signalr", func() bindings.OutputBinding {
						return signalr.NewSignalR(logContrib)
					}),
					bindings_loader.NewOutput("twilio.sms", func() bindings.OutputBinding {
						return sms.NewSMS(logContrib)
					}),
					bindings_loader.NewOutput("twilio.sendgrid", func() bindings.OutputBinding {
						return sendgrid.NewSendGrid(logContrib)
					}),
					bindings_loader.NewOutput("azure.eventgrid", func() bindings.OutputBinding {
						return eventgrid.NewAzureEventGrid(logContrib)
					}),
					bindings_loader.NewOutput("cron", func() bindings.OutputBinding {
						return cron.NewCron(logContrib)
					}),
					bindings_loader.NewOutput("twitter", func() bindings.OutputBinding {
						return twitter.NewTwitter(logContrib)
					}),
					bindings_loader.NewOutput("influx", func() bindings.OutputBinding {
						return influx.NewInflux(logContrib)
					}),
					bindings_loader.NewOutput("postgres", func() bindings.OutputBinding {
						return postgres.NewPostgres(logContrib)
					}),
					bindings_loader.NewOutput("postmark", func() bindings.OutputBinding {
						return postmark.NewPostmark(logContrib)
					}),
				),
				runtime.WithHTTPMiddleware(
					http_middleware_loader.New("uppercase", func(metadata middleware.Metadata) http_middleware.Middleware {
						return func(h fasthttp.RequestHandler) fasthttp.RequestHandler {
							return func(ctx *fasthttp.RequestCtx) {
								body := string(ctx.PostBody())
								ctx.Request.SetBody([]byte(strings.ToUpper(body)))
								h(ctx)
							}
						}
					}),
					http_middleware_loader.New("oauth2", func(metadata middleware.Metadata) http_middleware.Middleware {
						handler, _ := oauth2.NewOAuth2Middleware().GetHandler(metadata)
						return handler
					}),
					http_middleware_loader.New("oauth2clientcredentials", func(metadata middleware.Metadata) http_middleware.Middleware {
						handler, _ := oauth2clientcredentials.NewOAuth2ClientCredentialsMiddleware(log).GetHandler(metadata)
						return handler
					}),
					http_middleware_loader.New("ratelimit", func(metadata middleware.Metadata) http_middleware.Middleware {
						handler, _ := ratelimit.NewRateLimitMiddleware(log).GetHandler(metadata)
						return handler
					}),
					http_middleware_loader.New("bearer", func(metadata middleware.Metadata) http_middleware.Middleware {
						handler, _ := bearer.NewBearerMiddleware(log).GetHandler(metadata)
						return handler
					}),
					http_middleware_loader.New("opa", func(metadata middleware.Metadata) http_middleware.Middleware {
						handler, _ := opa.NewMiddleware(log).GetHandler(metadata)
						return handler
					}),
				),
			)
			if err != nil {
				log.Fatalf("fatal error from runtime: %s", err)
			}
			go startMosn(configPath, serviceCluster, serviceNode, serviceType, serviceMeta, serviceLabels, clusterDomain, podName, podNamespace, podIp, flagLogLevel, featureGates)

			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
			<-stop
			gracefulShutdownDuration := 5 * time.Second
			log.Info("mecha shutting down. Waiting 5 seconds to finish outstanding operations")
			rt.Stop()
			<-time.After(gracefulShutdownDuration)
			return nil
		},
	}

	cmdStop = cli.Command{
		Name:  "stop",
		Usage: "stop mecha sidecar",
		Action: func(c *cli.Context) error {
			return nil
		},
	}

	cmdReload = cli.Command{
		Name:  "reload",
		Usage: "reconfiguration",
		Action: func(c *cli.Context) error {
			return nil
		},
	}
)

func startMosn(configPath, serviceCluster, serviceNode, serviceType string, serviceMeta,
	metaLabels []string, clusterDomain, podName, podNamespace, podIp, flagLogLevel, featureGates string) {

	conf := configmanager.Load(configPath)
	if mosnLogLevel, ok := flagToMosnLogLevel[flagLogLevel]; ok {
		if mosnLogLevel == "OFF" {
			mosnlog.GetErrorLoggerManagerInstance().Disable()
		} else {
			mosnlog.GetErrorLoggerManagerInstance().SetLogLevelControl(configmanager.ParseLogLevel(mosnLogLevel))
		}
	}

	// set feature gates
	err := featuregate.Set(featureGates)
	if err != nil {
		mosnlog.StartLogger.Infof("[mosn] [start] parse feature-gates flag fail : %+v", err)
		os.Exit(1)
	}
	// start pprof
	if conf.Debug.StartDebug {
		port := 9090 //default use 9090
		if conf.Debug.Port != 0 {
			port = conf.Debug.Port
		}
		addr := fmt.Sprintf("0.0.0.0:%d", port)
		s := &rawhttp.Server{Addr: addr, Handler: nil}
		store.AddService(s, "pprof", nil, nil)
	}

	// set mosn metrics flush
	mosnmetrics.FlushMosnMetrics = true
	// set version and go version
	mosnmetrics.SetVersion(Version)
	mosnmetrics.SetGoVersion(rawruntime.Version())

	if serviceNode != "" {
		types.InitXdsFlags(serviceCluster, serviceNode, serviceMeta, metaLabels)
	} else {
		if types.IsApplicationNodeType(serviceType) {
			sn := podName + "." + podNamespace
			serviceNode := serviceType + "~" + podIp + "~" + sn + "~" + clusterDomain
			types.InitXdsFlags(serviceCluster, serviceNode, serviceMeta, metaLabels)
		} else {
			mosnlog.StartLogger.Infof("[mosn] [start] xds service type must be sidecar or router")
		}
	}

	mosn.Start(conf)
}
