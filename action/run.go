package action

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/BimaAdi/surl/core"
	"github.com/BimaAdi/surl/response"
	"github.com/urfave/cli/v3"
)

var _ = regexp.MustCompile(`\{([a-zA-Z0-9_.-]+)\}`)

func RunAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return errors.New("run requires exactly one request key")
	}

	output := cmd.String("output")
	if output != "json" && output != "raw" {
		return fmt.Errorf("invalid output %q: must be json or raw", output)
	}
	verbose := cmd.Bool("verbose")

	confPath := cmd.String("conf")
	if confPath == "" {
		confPath = cmd.Root().String("conf")
	}
	cfg, err := core.LoadConfig(confPath)
	if err != nil {
		return err
	}
	reqConfig, ok := cfg.API[cmd.Args().First()]
	if !ok {
		return fmt.Errorf("request %q not found in %s", cmd.Args().First(), confPath)
	}

	reqConfig, cfg, err = core.SubstituteRequest(reqConfig, cfg)
	if err != nil {
		return err
	}

	result, err := core.ExecuteRequest(ctx, http.DefaultClient, reqConfig, cfg)
	if err != nil {
		return err
	}
	var encoded []byte
	if output == "json" {
		encoded, err = response.ResponseJSON(result, verbose)
		if err != nil {
			return err
		}
		encoded = append(encoded, '\n')
	} else {
		encoded, err = response.RawResponse(result, verbose)
		if err != nil {
			return err
		}
	}
	_, err = cmd.Root().Writer.Write(encoded)
	return err
}
