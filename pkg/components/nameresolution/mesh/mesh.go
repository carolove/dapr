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
	mosnctx "mosn.io/mosn/pkg/context"
	"mosn.io/mosn/pkg/log"
	"mosn.io/mosn/pkg/protocol"
	"mosn.io/mosn/pkg/router"
	"mosn.io/mosn/pkg/types"
	"mosn.io/mosn/pkg/upstream/cluster"
)

type resolver struct {
	logger logger.Logger
}
const (
	DaprMeshRouterConfigName = "20880"
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
	headers := protocol.CommonHeader{}
	headers.Add(protocol.MosnHeaderHostKey, req.ID)
	context := mosnctx.WithValue(context.Background(), types.ContextKeyListenerName, req.ID)

	routersWrapper := router.GetRoutersMangerInstance().GetRouterWrapperByName(DaprMeshRouterConfigName)
	clusterManager := cluster.GetClusterMngAdapterInstance().ClusterManager
	routeHandlerFactory := router.GetMakeHandlerFunc(types.DefaultRouteHandler)

	// get router instance and do routing
	routers := routersWrapper.GetRouters()
	// call route handler to get route info
	snapshot, route := routeHandlerFactory.DoRouteHandler(context, headers, routers, clusterManager)
	cluster := snapshot.ClusterInfo()
	// as ClusterName has random factor when choosing weighted cluster,
	// so need determination at the first time
	clusterName := route.RouteRule().ClusterName()

	log.Proxy.Debugf(context, "[proxy] [downstream] route match result:%+v, clusterName=%v, cluster:=%v", route, clusterName,cluster)

	return fmt.Sprintf("%s-dapr.%s.svc.cluster.local:%d", req.ID, req.Namespace, req.Port), nil
}