package mock

import (
	"context"
	"fmt"
	"strings"

	"github.com/segmentio/ksuid"
	"github.com/spf13/cobra"
)

type StartOptions struct {
	addr     string
	protocol string
}

func runSrv(ctx context.Context, opts *StartOptions) error {
	srv := &ServerDemo{}
	srv.Start("srv1", opts.protocol, opts.addr)
	return nil
}

func runCli(ctx context.Context, opts *StartOptions) error {
	cli := &ClientDemo{}
	if opts.protocol == "ws" && !strings.HasPrefix(opts.addr, "ws:") {
		opts.addr = fmt.Sprintf("ws://%s", opts.addr)
	}
	cli.Start(ksuid.New().String(), opts.protocol, opts.addr)
	return nil
}

func NewServerCmd(ctx context.Context) *cobra.Command {
	opts := &StartOptions{}
	cmd := &cobra.Command{
		Use:   "mock_srv",
		Short: "start server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSrv(ctx, opts)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.addr, "address", "a", ":8000", "listen address")
	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol of tcp or ws")
	return cmd
}

func NewClientCmd(ctx context.Context) *cobra.Command {
	opts := &StartOptions{}
	cmd := &cobra.Command{
		Use:   "mock_cli",
		Short: "start client",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCli(ctx, opts)
		},
	}
	cmd.PersistentFlags().StringVarP(&opts.addr, "address", "a", "localhost:8000", "server address")
	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol of tcp or ws")
	return cmd
}
