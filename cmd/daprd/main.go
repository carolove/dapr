// ------------------------------------------------------------
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
// ------------------------------------------------------------

package main

import (
	"github.com/urfave/cli"
	"os"
	"strconv"
	"time"

	"github.com/dapr/dapr/pkg/logger"
	_ "mosn.io/mosn/pkg/admin/debug"
	_ "mosn.io/mosn/pkg/buffer"
	_ "mosn.io/mosn/pkg/filter/listener/originaldst"
	_ "mosn.io/mosn/pkg/filter/network/connectionmanager"
	_ "mosn.io/mosn/pkg/filter/network/proxy"
	_ "mosn.io/mosn/pkg/filter/network/streamproxy"
	_ "mosn.io/mosn/pkg/filter/stream/dsl"
	_ "mosn.io/mosn/pkg/filter/stream/dubbo"
	_ "mosn.io/mosn/pkg/filter/stream/faultinject"
	_ "mosn.io/mosn/pkg/filter/stream/faulttolerance"
	_ "mosn.io/mosn/pkg/filter/stream/flowcontrol"
	_ "mosn.io/mosn/pkg/filter/stream/gzip"
	_ "mosn.io/mosn/pkg/filter/stream/jwtauthn"
	_ "mosn.io/mosn/pkg/filter/stream/mirror"
	_ "mosn.io/mosn/pkg/filter/stream/mixer"
	_ "mosn.io/mosn/pkg/filter/stream/payloadlimit"
	_ "mosn.io/mosn/pkg/filter/stream/stats"
	_ "mosn.io/mosn/pkg/filter/stream/transcoder/http2bolt"
	_ "mosn.io/mosn/pkg/metrics/sink"
	_ "mosn.io/mosn/pkg/metrics/sink/prometheus"
	_ "mosn.io/mosn/pkg/network"
	_ "mosn.io/mosn/pkg/protocol"
	_ "mosn.io/mosn/pkg/protocol/http/conv"
	_ "mosn.io/mosn/pkg/protocol/http2/conv"
	_ "mosn.io/mosn/pkg/protocol/xprotocol"
	_ "mosn.io/mosn/pkg/protocol/xprotocol/bolt"
	_ "mosn.io/mosn/pkg/protocol/xprotocol/boltv2"
	_ "mosn.io/mosn/pkg/protocol/xprotocol/dubbo"
	_ "mosn.io/mosn/pkg/protocol/xprotocol/tars"
	_ "mosn.io/mosn/pkg/router"
	_ "mosn.io/mosn/pkg/stream/http"
	_ "mosn.io/mosn/pkg/stream/http2"
	_ "mosn.io/mosn/pkg/stream/xprotocol"
	_ "mosn.io/mosn/pkg/trace/skywalking"
	_ "mosn.io/mosn/pkg/trace/skywalking/http"
	_ "mosn.io/mosn/pkg/trace/sofa/http"
	_ "mosn.io/mosn/pkg/trace/sofa/xprotocol"
	_ "mosn.io/mosn/pkg/trace/sofa/xprotocol/bolt"
	_ "mosn.io/mosn/pkg/upstream/healthcheck"
	_ "mosn.io/mosn/pkg/upstream/servicediscovery/dubbod"
	_ "mosn.io/mosn/pkg/xds"
	// Included components in compiled daprd
)

var (
	log        = logger.NewLogger("dapr.runtime")
	logContrib = logger.NewLogger("dapr.contrib")

	// Version dapr version
	Version = "1.0.0-rc.2"
)

func main() {
	app := newDaprApp(&cmdStart)

	// ignore error so we don't exit non-zero and break gfmrun README example tests
	_ = app.Run(os.Args)
}

func newDaprApp(startCmd *cli.Command) *cli.App {
	app := cli.NewApp()
	app.Name = "mecha"
	app.Version = Version
	app.Compiled = time.Now()
	app.Copyright = "(c) " + strconv.Itoa(time.Now().Year()) + " JD Group"
	app.Usage = "mecha"
	app.Flags = cmdStart.Flags

	//commands
	app.Commands = []cli.Command{
		cmdStart,
		cmdStop,
		cmdReload,
	}

	//action
	app.Action = func(c *cli.Context) error {
		if c.NumFlags() == 0 {
			return cli.ShowAppHelp(c)
		}

		return startCmd.Action.(func(c *cli.Context) error)(c)
	}

	return app
}
