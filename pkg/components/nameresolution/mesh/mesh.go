package mesh

// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

import (
	"context"
	"fmt"

	"github.com/dapr/components-contrib/nameresolution"
	"github.com/dapr/dapr/pkg/logger"
	"mosn.io/mosn/pkg/log"
	"mosn.io/mosn/pkg/protocol"
	"mosn.io/mosn/pkg/router"
	"mosn.io/mosn/pkg/types"
	"mosn.io/mosn/pkg/upstream/cluster"
	"mosn.io/mosn/pkg/variable"
)

type resolver struct {
	logger logger.Logger
}

const (
	DaprMeshRouterConfigName = "20882"
)

var (
	NamingResolverNoRouteErr = fmt.Errorf("[proxy] RouterConfigName:%s doesn't exit", DaprMeshRouterConfigName)
)

// NewResolver creates Kubernetes name resolver.
func NewResolver(logger logger.Logger) nameresolution.Resolver {
	return &resolver{logger: logger}
}

// Init initializes Kubernetes name resolver.
func (k *resolver) Init(metadata nameresolution.Metadata) error {
	return nil
}

// ResolveID resolves name to address in Mesh.
func (k *resolver) ResolveID(req nameresolution.ResolveRequest) (string, error) {

	headers := protocol.CommonHeader(map[string]string{
		protocol.MosnHeaderHostKey: "mosn.io.dubbo.proxy",
	})
	context := variable.NewVariableContext(context.Background())

	// adapte dubbo service to http host
	variable.SetVariableValue(context, protocol.MosnHeaderHostKey, "mosn.io.dubbo.proxy")
	// because use http rule, so should add default path
	variable.SetVariableValue(context, protocol.MosnHeaderPathKey, "/")

	routersWrapper := router.GetRoutersMangerInstance().GetRouterWrapperByName(DaprMeshRouterConfigName)
	clusterManager := cluster.GetClusterMngAdapterInstance().ClusterManager
	routeHandlerFactory := router.GetMakeHandlerFunc(types.DefaultRouteHandler)

	// get router instance and do routing
	routers := routersWrapper.GetRouters()

	// log.Proxy.Warnf(context, "[proxy] [downstream] routes:%+v, ", routers)
	// time.Sleep(2 * time.Second)
	// routes:&{virtualHostsIndex:map[mosn.io.dubbo.proxy:0 mosn.io.dubbo.proxy:20882:0] defaultVirtualHostIndex:1 wildcardVirtualHostSuffixesIndex:map[] greaterSortedWildcardVirtualHostSuffixes:[] virtualHosts:[0xc0014d6d80 0xc0014d6de0]},
	// 已经找到了对应的route了

	// call route handler to get route info
	snapshot, route := routeHandlerFactory.DoRouteHandler(context, headers, routers, clusterManager)

	// log.Proxy.Warnf(context, "[proxy] [downstream] snapshot:%+v, route:%+v", snapshot,route)
	// time.Sleep(2 * time.Second)
	//  snapshot:&{info:0xc001bef680 hostSet:0xc001bf46c0 lb:0xc001bfa2a0}, route:&{RouteRuleImplBase:0xc001c94780 prefix:/}

	cluster := snapshot.ClusterInfo()

	// as ClusterName has random factor when choosing weighted cluster,
	// so need determination at the first time
	clusterName := route.RouteRule().ClusterName()

	log.Proxy.Warnf(context, "[proxy] [downstream] route match result:%+v, clusterName=%v, cluster:=%v, req=%#v", route, clusterName, cluster, req)
	hosts := snapshot.HostSet().Hosts()
	if len(hosts) != 0 {
		host := hosts[0]
		log.Proxy.Warnf(context, "[proxy] [downstream] route match result:%+v, clusterName=%v, cluster:=%v, host:%#v, addr:%s", route, clusterName, cluster, host, host.AddressString())
		return host.AddressString(), nil
	}

	return fmt.Sprintf("%s-dapr.%s.svc.cluster.local:%d", req.ID, req.Namespace, req.Port), nil
}
