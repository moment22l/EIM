package router

import (
	"context"

	"github.com/kataras/iris/v12"
	"github.com/spf13/cobra"
)

type ServerStartOptions struct {
	config string
}

// RunServerStart 启动路由服务器
func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	app := iris.Default()
	app.Get("/health", func(ctx iris.Context) {
		_, _ = ctx.WriteString("ok")
	})
	// 开启服务器
	return app.Listen("", iris.WithOptimizations)
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "router",
		Short: "start a router",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./router/conf.yaml", "Config file")
	return cmd
}
