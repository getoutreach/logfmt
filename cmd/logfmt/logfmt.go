// Copyright 2026 Outreach Corporation. Licensed under the Apache License 2.0.

// Description: This file is the entrypoint for the logfmt CLI
// command for logfmt.
// Managed: true

package main

import (
	"context"

	oapp "github.com/getoutreach/gobox/pkg/app"
	"github.com/getoutreach/gobox/pkg/cfg"
	gcli "github.com/getoutreach/gobox/pkg/cli"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"

	// Place any extra imports for your startup code here
	// <<Stencil::Block(imports)>>
	"github.com/getoutreach/logfmt/internal/runner"
	// <</Stencil::Block>>
)

// HoneycombTracingKey gets set by the Makefile at compile-time which is pulled
// down by devconfig.sh.
var HoneycombTracingKey = "NOTSET" //nolint:gochecknoglobals // Why: We can't compile in things as a const.

// TeleforkAPIKey gets set by the Makefile at compile-time which is pulled
// down by devconfig.sh.
var TeleforkAPIKey = "NOTSET" //nolint:gochecknoglobals // Why: We can't compile in things as a const.

// <<Stencil::Block(honeycombDataset)>>

// HoneycombDataset is a constant denoting the dataset that traces should be stored
// in in honeycomb.
const HoneycombDataset = ""

// <</Stencil::Block>>

// <<Stencil::Block(global)>>

// <</Stencil::Block>>

// main is the entrypoint for the logfmt CLI.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	log := logrus.New()

	// <<Stencil::Block(init)>>

	// <</Stencil::Block>>

	app := cli.Command{
		Version:               oapp.Version,
		Name:                  "logfmt",
		EnableShellCompletion: true,
		// <<Stencil::Block(app)>>
		Usage: "Formats structured logs to a human readable format.",
		Description: `Example:

    make test | logfmt --filter <filter> --format <format>`,
		Action: func(ctx context.Context, c *cli.Command) error {
			r := runner.New(log, c.String("filter"), c.String("format"))
			r.Run()
			return nil
		},
		// <</Stencil::Block>>
	}
	app.Flags = []cli.Flag{
		// <<Stencil::Block(flags)>>
		&cli.StringFlag{
			Name:  "filter",
			Usage: "filter the log. Use jq syntax",
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "format the output.  Use golang templates syntax",
		},
		// <</Stencil::Block>>
	}
	app.Commands = []*cli.Command{
		// <<Stencil::Block(commands)>>

		// <</Stencil::Block>>
	}

	// <<Stencil::Block(postApp)>>

	// <</Stencil::Block>>

	// Insert global flags, tracing, updating and start the application.
	gcli.RunV3(ctx, cancel, &app, &gcli.Config{
		Logger: log,
		Telemetry: gcli.TelemetryConfig{
			Otel: gcli.TelemetryOtelConfig{
				Dataset:         HoneycombDataset,
				HoneycombAPIKey: cfg.SecretData(HoneycombTracingKey),
			},
		},
	})
}
