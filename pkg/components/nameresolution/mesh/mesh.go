// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package mesh

import (
	"github.com/dapr/components-contrib/nameresolution"
	"github.com/dapr/dapr/pkg/logger"
)

type resolver struct {
	logger logger.Logger
}

// NewResolver creates Kubernetes Mesh name resolver.
func NewResolver(logger logger.Logger) nameresolution.Resolver {
	return &resolver{logger: logger}
}

// Init initializes Kubernetes Mesh name resolver.
func (k *resolver) Init(metadata nameresolution.Metadata) error {
	return nil
}

// ResolveID resolves name to address in Kubernetes Mesh.
func (k *resolver) ResolveID(req nameresolution.ResolveRequest) (string, error) {
	// Dapr requires this formatting for Kubernetes Mesh services
	return req.ID, nil
}
