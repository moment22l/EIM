package main

import (
	"EIM/examples/mock"
	"EIM/logger"
	"context"
	"flag"

	"github.com/spf13/cobra"
)

const version = "v1"

func main() {
	flag.Parse()
	root := &cobra.Command{
		Use:     "EIM",
		Version: version,
		Short:   "server",
	}
	ctx := context.Background()
	root.AddCommand(mock.NewServerCmd(ctx))
	root.AddCommand(mock.NewClientCmd(ctx))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
