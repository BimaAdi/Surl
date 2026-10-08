package action

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/BimaAdi/surl/core"
	"github.com/urfave/cli/v3"
)

var variablePattern = regexp.MustCompile(`\{([a-zA-Z0-9_.-]+)\}`)

func RunAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() != 1 {
		return errors.New("run requires exactly one request key")
	}

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
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("encode response: %w", err)
	}
	encoded = append(encoded, '\n')
	_, err = cmd.Root().Writer.Write(encoded)
	return err
}
