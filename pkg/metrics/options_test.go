// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package metrics

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	t.Run("default options", func(t *testing.T) {
		o := defaultMetricOptions()
		assert.Equal(t, DefaultMetricsPort, o.MetricsPort)
		assert.Equal(t, DefaultMetricsEnabled, o.MetricsEnabled)
	})

	t.Run("attaching metrics related cmd flags", func(t *testing.T) {
		o := defaultMetricOptions()

		metricsPortAsserted := false
		testStringVarFn := func(p *string, name string, value string, usage string) {
			if name == "metrics-port" && value == DefaultMetricsPort {
				metricsPortAsserted = true
			}
		}

		metricsEnabledAsserted := false
		testBoolVarFn := func(p *bool, name string, value bool, usage string) {
			if name == "enable-metrics" && value == DefaultMetricsEnabled {
				metricsEnabledAsserted = true
			}
		}

		o.AttachCmdFlags(testStringVarFn, testBoolVarFn)

		// assert
		assert.True(t, metricsPortAsserted)
		assert.True(t, metricsEnabledAsserted)
	})

	t.Run("parse valid port", func(t *testing.T) {
		o := Options{
			MetricsPort:    "1010",
			MetricsEnabled: false,
		}

		assert.Equal(t, uint64(1010), o.GetMetricsPort())
	})

	t.Run("return default port if port is invalid", func(t *testing.T) {
		o := Options{
			MetricsPort:    "invalid",
			MetricsEnabled: false,
		}

		defaultPort, _ := strconv.ParseUint(DefaultMetricsPort, 10, 64)

		assert.Equal(t, defaultPort, o.GetMetricsPort())
	})

	t.Run("attaching single metrics related cmd flag", func(t *testing.T) {
		o := defaultMetricOptions()

		metricsPortAsserted := false
		testStringVarFn := func(p *string, name string, value string, usage string) {
			if name == "metrics-port" && value == DefaultMetricsPort {
				metricsPortAsserted = true
			}
		}

		o.AttachCmdFlag(testStringVarFn)

		// assert
		assert.True(t, metricsPortAsserted)
	})
}
