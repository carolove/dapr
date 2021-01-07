GetRouterWrapperByName(routerConfigName string) RouterWrapper
这个函数可以获取到route name相关的route的所有信息，如果为空则说明xds rds关于这个route的数据没有获取到
这一步发生在proxy初始化过程中
其次还需要把route routeHandlerFactory也初始化
proxy.routeHandlerFactory = router.GetMakeHandlerFunc(proxy.config.RouterHandlerName)，应该是default route factory
最后还需要把cluster manager的instance也获取到


// get router instance and do routing
	routers := s.proxy.routersWrapper.GetRouters()
这一步发生在downstream 的get routers过程中，获取所有的routers数据以便进行match

s.snapshot, s.route = s.proxy.routeHandlerFactory.DoRouteHandler(s.context, headers, routers, s.proxy.clusterManager)
获取到了对应的route和snapshot

```yaml
---
apiVersion: networking.istio.io/v1beta1
kind: WorkloadEntry
metadata:
  name: proxy-2
spec:
  address: 10.1.0.136
  labels:
    app: mosn.io.dubbo.proxy

---
apiVersion: networking.istio.io/v1beta1
kind: ServiceEntry
metadata:
  name: dubbo-proxy
spec:
  hosts:
    - mosn.io.dubbo.proxy
  ports:
    - number: 20882
      name: dubbo-http
      protocol: HTTP
	  targetPort: 50002
  location: MESH_INTERNAL
  resolution: STATIC
  workloadSelector:
    labels:
      app: mosn.io.dubbo.proxy
```

```yaml
kind: Service
apiVersion: v1
metadata:
  name: nodeapp
  labels:
    app: nodeapp
spec:
  selector:
    app: nodeapp
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3000
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: nodeapp
  name: nodeapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: nodeapp
  strategy: {}
  template:
    metadata:
      annotations:
        dapr.io/app-id: nodeapp
        dapr.io/app-port: "3000"
        dapr.io/enabled: "true"
        prometheus.io/path: /stats/prometheus
        prometheus.io/port: "15020"
        prometheus.io/scrape: "true"
        sidecar.istio.io/capNetBindService: "true"
        sidecar.istio.io/inject: "true"
        sidecar.istio.io/interceptionMode: NONE
        sidecar.istio.io/proxyImage: docker.io/mosnio/mosn:1.7.5-mosn
        sidecar.istio.io/status: '{"version":"cab296262ce64bc850d7469f07dea93008e82d0c75ebae831dea2d3fe0bc6de9","initContainers":null,"containers":["istio-proxy"],"volumes":["istio-envoy","istio-data","istio-podinfo","istiod-ca-cert"],"imagePullSecrets":null}'
        traffic.sidecar.istio.io/excludeInboundPorts: "15020"
        traffic.sidecar.istio.io/includeInboundPorts: "3000"
        traffic.sidecar.istio.io/includeOutboundIPRanges: '*'
      creationTimestamp: null
      labels:
        app: nodeapp
        istio.io/rev: ""
        security.istio.io/tlsMode: istio
    spec:
      containers:
      - env:
        - name: DAPR_HTTP_PORT
          value: "3500"
        - name: DAPR_GRPC_PORT
          value: "50001"
        image: dapriosamples/hello-k8s-node:latest
        imagePullPolicy: Always
        name: node
        ports:
        - containerPort: 3000
        resources: {}
      - args:
        - proxy
        - sidecar
        - --mode
        - kubernetes
        - --dapr-http-port
        - "3500"
        - --dapr-grpc-port
        - "50001"
        - --dapr-internal-grpc-port
        - "50002"
        - --app-port
        - "3000"
        - --app-id
        - nodeapp
        - --control-plane-address
        - dapr-api.dapr-system.svc.cluster.local:80
        - --app-protocol
        - http
        - --placement-host-address
        - dapr-placement-server.dapr-system.svc.cluster.local:50005
        - --config
        - ""
        - --log-level
        - "info"
        - --app-max-concurrency
        - "-1"
        - --sentry-address
        - dapr-sentry.dapr-system.svc.cluster.local:80
        - --metrics-port
        - "9090"
        - --domain
        - $(POD_NAMESPACE).svc.cluster.local
        - --serviceCluster
        - nodeapp.$(POD_NAMESPACE)
        - --proxyLogLevel=info
        - --proxyComponentLogLevel=misc:error
        - --trust-domain=cluster.local
        - --concurrency
        - "2"
        env:
        - name: JWT_POLICY
          value: first-party-jwt
        - name: PILOT_CERT_PROVIDER
          value: istiod
        - name: CA_ADDR
          value: istiod.istio-system.svc:15012
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: DAPR_HOST_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        - name: SERVICE_ACCOUNT
          valueFrom:
            fieldRef:
              fieldPath: spec.serviceAccountName
        - name: HOST_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        - name: CANONICAL_SERVICE
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['service.istio.io/canonical-name']
        - name: CANONICAL_REVISION
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['service.istio.io/canonical-revision']
        - name: PROXY_CONFIG
          value: |
            {"binaryPath":"/usr/local/bin/mosn","customConfigFile":"/etc/istio/mosn/mosn_config_dubbo_xds.json","proxyMetadata":{"DNS_AGENT":""},"statusPort":15021}
        - name: ISTIO_META_POD_PORTS
          value: |-
            [
                {"containerPort":3000}
            ]
        - name: ISTIO_META_APP_CONTAINERS
          value: node
        - name: ISTIO_META_CLUSTER_ID
          value: Kubernetes
        - name: ISTIO_META_INTERCEPTION_MODE
          value: NONE
        - name: ISTIO_METAJSON_ANNOTATIONS
          value: |
            {"dapr.io/app-id":"nodeapp","dapr.io/app-port":"3000","dapr.io/enabled":"true","sidecar.istio.io/capNetBindService":"true","sidecar.istio.io/inject":"true","sidecar.istio.io/interceptionMode":"NONE","sidecar.istio.io/proxyImage":"docker.io/mosnio/mosn:1.7.5-mosn"}
        - name: ISTIO_META_WORKLOAD_NAME
          value: nodeapp
        - name: ISTIO_META_OWNER
          value: kubernetes://apis/apps/v1/namespaces/default/deployments/nodeapp
        - name: ISTIO_META_MESH_ID
          value: cluster.local
        - name: DNS_AGENT
        - name: ISTIO_KUBE_APP_PROBERS
          value: '{}'
        image: docker.io/mosnio/mosn:1.7.5-mosn
        imagePullPolicy: IfNotPresent
        name: istio-proxy
        ports:
        - containerPort: 15090
          name: http-envoy-prom
          protocol: TCP
        - containerPort: 3500
          name: dapr-http
          protocol: TCP
        - containerPort: 50001
          name: dapr-grpc
          protocol: TCP
        - containerPort: 50002
          name: dapr-internal
          protocol: TCP
        - containerPort: 9090
          name: dapr-metrics
          protocol: TCP
        readinessProbe:
          failureThreshold: 30
          httpGet:
            path: /healthz/ready
            port: 15021
          initialDelaySeconds: 1
          periodSeconds: 2
        resources:
          limits:
            cpu: "2"
            memory: 1Gi
          requests:
            cpu: 100m
            memory: 128Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - ALL
          privileged: false
          readOnlyRootFilesystem: true
          runAsGroup: 1337
          runAsNonRoot: false
          runAsUser: 0
        volumeMounts:
        - mountPath: /var/run/secrets/istio
          name: istiod-ca-cert
        - mountPath: /var/lib/istio/data
          name: istio-data
        - mountPath: /etc/istio/proxy
          name: istio-envoy
        - mountPath: /etc/istio/pod
          name: istio-podinfo
      securityContext:
        fsGroup: 1337
      volumes:
      - emptyDir:
          medium: Memory
        name: istio-envoy
      - emptyDir: {}
        name: istio-data
      - downwardAPI:
          items:
          - fieldRef:
              fieldPath: metadata.labels
            path: labels
          - fieldRef:
              fieldPath: metadata.annotations
            path: annotations
        name: istio-podinfo
      - configMap:
          name: istio-ca-root-cert
        name: istiod-ca-cert
status: {}
---
```

```yaml
kind: Service
apiVersion: v1
metadata:
  name: pythonapp
  labels:
    app: pythonapp
spec:
  selector:
    app: pythonapp
  ports:
  - protocol: TCP
    port: 80
    targetPort: 3000
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: pythonapp
  name: pythonapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pythonapp
  strategy: {}
  template:
    metadata:
      annotations:
        dapr.io/app-id: pythonapp
        dapr.io/enabled: "true"
        prometheus.io/path: /stats/prometheus
        prometheus.io/port: "15020"
        prometheus.io/scrape: "true"
        sidecar.istio.io/capNetBindService: "true"
        sidecar.istio.io/inject: "true"
        sidecar.istio.io/interceptionMode: NONE
        sidecar.istio.io/proxyImage: docker.io/mosnio/mosn:1.7.5-mosn
        sidecar.istio.io/status: '{"version":"cab296262ce64bc850d7469f07dea93008e82d0c75ebae831dea2d3fe0bc6de9","initContainers":null,"containers":["istio-proxy"],"volumes":["istio-envoy","istio-data","istio-podinfo","istiod-ca-cert"],"imagePullSecrets":null}'
        traffic.sidecar.istio.io/excludeInboundPorts: "15020"
        traffic.sidecar.istio.io/includeInboundPorts: "3000"
        traffic.sidecar.istio.io/includeOutboundIPRanges: '*'
      creationTimestamp: null
      labels:
        app: pythonapp
        istio.io/rev: ""
        security.istio.io/tlsMode: istio
    spec:
      containers:
      - env:
        - name: DAPR_HTTP_PORT
          value: "3500"
        - name: DAPR_GRPC_PORT
          value: "50001"
        image: dapriosamples/hello-k8s-python:latest
        name: python
        ports:
        - containerPort: 3000
        resources: {}
      - args:
        - proxy
        - sidecar
        - --mode
        - kubernetes
        - --dapr-http-port
        - "3500"
        - --dapr-grpc-port
        - "50001"
        - --dapr-internal-grpc-port
        - "50002"
        - --app-port
        - ""
        - --app-id
        - pythonapp
        - --control-plane-address
        - dapr-api.dapr-system.svc.cluster.local:80
        - --app-protocol
        - http
        - --placement-host-address
        - dapr-placement-server.dapr-system.svc.cluster.local:50005
        - --config
        - ""
        - --log-level
        - "info"
        - --app-max-concurrency
        - "-1"
        - --sentry-address
        - dapr-sentry.dapr-system.svc.cluster.local:80
        - --metrics-port
        - "9090"
        - --domain
        - $(POD_NAMESPACE).svc.cluster.local
        - --serviceCluster
        - pythonapp.$(POD_NAMESPACE)
        - --proxyLogLevel=info
        - --proxyComponentLogLevel=misc:error
        - --trust-domain=cluster.local
        - --concurrency
        - "2"
        env:
        - name: JWT_POLICY
          value: first-party-jwt
        - name: PILOT_CERT_PROVIDER
          value: istiod
        - name: CA_ADDR
          value: istiod.istio-system.svc:15012
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: INSTANCE_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        - name: DAPR_HOST_IP
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: status.podIP
        - name: SERVICE_ACCOUNT
          valueFrom:
            fieldRef:
              fieldPath: spec.serviceAccountName
        - name: HOST_IP
          valueFrom:
            fieldRef:
              fieldPath: status.hostIP
        - name: CANONICAL_SERVICE
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['service.istio.io/canonical-name']
        - name: CANONICAL_REVISION
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['service.istio.io/canonical-revision']
        - name: PROXY_CONFIG
          value: |
            {"binaryPath":"/usr/local/bin/mosn","customConfigFile":"/etc/istio/mosn/mosn_config_dubbo_xds.json","proxyMetadata":{"DNS_AGENT":""},"statusPort":15021}
        - name: ISTIO_META_POD_PORTS
          value: |-
            [
                {"containerPort":3000}
            ]
        - name: ISTIO_META_APP_CONTAINERS
          value: python
        - name: ISTIO_META_CLUSTER_ID
          value: Kubernetes
        - name: ISTIO_META_INTERCEPTION_MODE
          value: NONE
        - name: ISTIO_METAJSON_ANNOTATIONS
          value: |
            {"dapr.io/app-id":"pythonapp","dapr.io/enabled":"true","sidecar.istio.io/capNetBindService":"true","sidecar.istio.io/inject":"true","sidecar.istio.io/interceptionMode":"NONE","sidecar.istio.io/proxyImage":"docker.io/mosnio/mosn:1.7.5-mosn"}
        - name: ISTIO_META_WORKLOAD_NAME
          value: pythonapp
        - name: ISTIO_META_OWNER
          value: kubernetes://apis/apps/v1/namespaces/default/deployments/pythonapp
        - name: ISTIO_META_MESH_ID
          value: cluster.local
        - name: DNS_AGENT
        - name: ISTIO_KUBE_APP_PROBERS
          value: '{}'
        image: docker.io/mosnio/mosn:1.7.5-mosn
        imagePullPolicy: IfNotPresent
        name: istio-proxy
        ports:
        - containerPort: 15090
          name: http-envoy-prom
          protocol: TCP
        - containerPort: 3500
          name: dapr-http
          protocol: TCP
        - containerPort: 50001
          name: dapr-grpc
          protocol: TCP
        - containerPort: 50002
          name: dapr-internal
          protocol: TCP
        - containerPort: 9090
          name: dapr-metrics
          protocol: TCP
        readinessProbe:
          failureThreshold: 30
          httpGet:
            path: /healthz/ready
            port: 15021
          initialDelaySeconds: 1
          periodSeconds: 2
        resources:
          limits:
            cpu: "2"
            memory: 1Gi
          requests:
            cpu: 100m
            memory: 128Mi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - ALL
          privileged: false
          readOnlyRootFilesystem: true
          runAsGroup: 1337
          runAsNonRoot: false
          runAsUser: 0
        volumeMounts:
        - mountPath: /var/run/secrets/istio
          name: istiod-ca-cert
        - mountPath: /var/lib/istio/data
          name: istio-data
        - mountPath: /etc/istio/proxy
          name: istio-envoy
        - mountPath: /etc/istio/pod
          name: istio-podinfo
      securityContext:
        fsGroup: 1337
      volumes:
      - emptyDir:
          medium: Memory
        name: istio-envoy
      - emptyDir: {}
        name: istio-data
      - downwardAPI:
          items:
          - fieldRef:
              fieldPath: metadata.labels
            path: labels
          - fieldRef:
              fieldPath: metadata.annotations
            path: annotations
        name: istio-podinfo
      - configMap:
          name: istio-ca-root-cert
        name: istiod-ca-cert
status: {}
---
```

```json
{
	"servers": [
		{
			"default_log_path": "stdout",
			"default_log_level": "DEBUG",
			"listeners": [
				{
					"name": "outbound_listener",
					"address": "0.0.0.0:20881",
					"bind_port": true,
					"access_logs": [
						{
							"log_path": "stdout",
							"log_format": "[%start_time%] %request_received_duration% %response_received_duration% %bytes_sent% %bytes_received% %protocol% %response_code% %duration% %response_flag% %response_code% %upstream_local_address% %downstream_local_address% %downstream_remote_address% %upstream_host% %upstream_transport_failure_reason% %upstream_cluster%"
						}
					],
					"filter_chains": [
						{
							"tls_context": {},
							"filters": [
								{
									"type": "proxy",
									"config": {
										"downstream_protocol": "X",
										"upstream_protocol": "X",
										"router_config_name": "20882",
										"extend_config": {
											"sub_protocol": "dubbo"
										}
									}
								}
							]
						}
					],
					"stream_filters": [
						{
							"type": "dubbo_stream",
							"config": {
                                "subset": "group"
							}
						}
					]
				},
				{
					"name": "inbound_listener",
					"address": "0.0.0.0:20882",
					"bind_port": true,
                    "log_path": "stdout",
					"filter_chains": [
						{
							"tls_context": {},
							"filters": [
								{
									"type": "proxy",
									"config": {
										"downstream_protocol": "X",
										"upstream_protocol": "X",
										"router_config_name": "dubbo_provider_router",
										"extend_config": {
											"sub_protocol": "dubbo"
										}
									}
								}
							]
						}
					],
					"stream_filters": [
						{
							"type": "dubbo_stream",
							"config": {
                                "subset": "group"
							}
						}
					]
				}
			],
			"routers": [
				{
					"router_config_name": "dubbo_provider_router",
					"virtual_hosts": [
						{
							"name": "provider",
							"domains": ["*"],
							"routers": [
								{
									"match": {
										"headers": [
											{
												"name": "service",
												"value": ".*"
											}
										]
									},
									"route": {
										"cluster_name": "dubbo_provider_cluster"
									}
								}
							]
						}
					]
				}
			]
		}
	],
	"cluster_manager": {
		"clusters": [
			{
				"name": "dubbo_provider_cluster",
				"type": "SIMPLE",
				"lb_type": "LB_RANDOM",
				"max_request_per_conn": 1024,
				"conn_buffer_limit_bytes": 32768,
				"hosts": [
					{
						"address": "127.0.0.1:20880"
					}
				]
			}
		]
	},
	"metrics": {
		"sinks": [
			{
				"type": "prometheus",
				"config": {
					"port": 15090,
					"endpoint": "/stats/prometheus",
                    "percentiles": [50, 90, 95, 96, 99]
				}
			}
		]
	},
	"admin": {
		"address": {
			"socket_address": {
				"address": "0.0.0.0",
				"port_value": "15000"
			}
		}
	},
	"static_resources": {
		"clusters": [
			{
				"name": "xds-grpc",
				"type": "STRICT_DNS",
				"respect_dns_ttl": true,
				"dns_lookup_family": "V4_ONLY",
				"connect_timeout": "1s",
				"lb_policy": "ROUND_ROBIN",
				"load_assignment": {
					"cluster_name": "xds-grpc",
					"endpoints": [
						{
							"lb_endpoints": [
								{
									"endpoint": {
										"address": {
											"socket_address": {
												"address": "istiod.istio-system.svc",
												"port_value": 15010
											}
										}
									}
								}
							]
						}
					]
				}
			}
		]
	},
	"dynamic_resources": {
		"lds_config": {
			"ads": {}
		},
		"cds_config": {
			"ads": {}
		},
		"ads_config": {
			"api_type": "GRPC",
			"grpc_services": [
				{
					"envoy_grpc": {
						"cluster_name": "xds-grpc"
					}
				}
			],
			"refresh_delay": {
				"seconds": 20
			}
		}
	}
}
```

```dockerfile
FROM istio/proxyv2:1.7.6

COPY build/mosn /usr/local/bin/mosn
COPY build/pilot-agent /usr/local/bin/pilot-agent
COPY build/mosn_config_dubbo_xds.json /etc/istio/mosn/mosn_config_dubbo_xds.json
```

