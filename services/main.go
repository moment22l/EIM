package main

import (
	"EIM/logger"
	"EIM/services/gateway"
	"EIM/services/router"
	"EIM/services/server"
	"EIM/services/service"
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
		Short:   "Elegant Instant Messenger Cloud",
	}
	ctx := context.Background()

	root.AddCommand(gateway.NewServerStartCmd(ctx, version))
	root.AddCommand(server.NewServerStartCmd(ctx, version))
	root.AddCommand(service.NewServerStartCmd(ctx, version))
	root.AddCommand(router.NewServerStartCmd(ctx, version))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
