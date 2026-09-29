package main

import (
	"context"
	"fmt"
	"os"

	"github.com/BimaAdi/surl/action"
	"github.com/BimaAdi/surl/core"
	"github.com/urfave/cli/v3"
)

// These aliases and delegates keep the existing main-package tests focused on
// the same API while the implementations live in core.
type config = core.Config
type requestConfig = core.RequestConfig
type response = core.Response

func main() {
	if err := newApp().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newApp() *cli.Command {
	return &cli.Command{
		Name:  "surl",
		Usage: "make HTTP requests from surl.json metadata",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "conf", Aliases: []string{"c"}, Value: "surl.json", Usage: "path to the configuration JSON file"},
		},
		Commands: []*cli.Command{{
			Name:      "run",
			Usage:     "run a request by its key",
			ArgsUsage: "<key>",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "conf", Aliases: []string{"c"}, Usage: "path to the configuration JSON file"},
			},
			Action: action.RunAction,
		}, {
			Name:  "list",
			Usage: "list the available request keys",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "conf", Aliases: []string{"c"}, Usage: "path to the configuration JSON file"},
			},
			Action: action.ListAction,
		}},
	}
}
