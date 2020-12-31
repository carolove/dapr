// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package metrics

import (
	"strconv"
)

const (
	DefaultMetricsPort    = "9090"
	DefaultMetricsEnabled = true
)

// Options defines the sets of options for Dapr logging
type Options struct {
	// OutputLevel is the level of logging
	MetricsEnabled bool

	MetricsPort string
}

func defaultMetricOptions() *Options {
	return &Options{
		MetricsPort:    DefaultMetricsPort,
		MetricsEnabled: DefaultMetricsEnabled,
	}
}

// GetMetricsPort gets metrics port.
func (o *Options) GetMetricsPort() uint64 {
	port, err := strconv.ParseUint(o.MetricsPort, 10, 64)
	if err != nil {
		// Use default metrics port as a fallback
		port, _ = strconv.ParseUint(DefaultMetricsPort, 10, 64)
	}

	return port
}

// AttachCmdFlags attaches metrics options to command flags
func (o *Options) AttachCmdFlags(
	stringVar func(p *string, name string, value string, usage string),
	boolVar func(p *bool, name string, value bool, usage string)) {
	stringVar(
		&o.MetricsPort,
		"metrics-port",
		DefaultMetricsPort,
		"The port for the metrics server")
	boolVar(
		&o.MetricsEnabled,
		"enable-metrics",
		DefaultMetricsEnabled,
		"Enable prometheus metric")
}

// AttachCmdFlag attaches single metrics option to command flags
func (o *Options) AttachCmdFlag(
	stringVar func(p *string, name string, value string, usage string)) {
	stringVar(
		&o.MetricsPort,
		"metrics-port",
		DefaultMetricsPort,
		"The port for the metrics server")
}
