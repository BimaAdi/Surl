package action

import (
	"context"
	"fmt"
	"sort"

	"github.com/BimaAdi/surl/core"
	"github.com/urfave/cli/v3"
)

func ListAction(ctx context.Context, cmd *cli.Command) error {
	confPath := cmd.String("conf")
	if confPath == "" {
		confPath = cmd.Root().String("conf")
	}
	cfg, err := core.LoadConfig(confPath)
	if err != nil {
		return err
	}

	apiNames := make([]string, 0, len(cfg.API))
	for name := range cfg.API {
		apiNames = append(apiNames, name)
	}
	sort.Strings(apiNames)
	for _, name := range apiNames {
		if _, err := fmt.Fprintln(cmd.Root().Writer, name); err != nil {
			return err
		}
	}
	return nil
}
