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
