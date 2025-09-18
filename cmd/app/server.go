package app

import (
	"awesome-alert/api/router"
	"awesome-alert/cmd/app/options"
	"awesome-alert/internal/ctx"
	"awesome-alert/internal/factory"
	"context"
	"github.com/spf13/cobra"
)

func NewServerCommand() *cobra.Command {
	opts := options.NewOptions()

	cmd := &cobra.Command{
		Use: "awesome-alert",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(); err != nil {
				return err
			}

			if err := opts.Register(); err != nil {
				return err
			}

			return run(opts)
		},
		Args: cobra.NoArgs,
	}

	opts.BindFlags(cmd)

	return cmd
}

func run(opts *options.Options) error {

	ctxObj := ctx.NewContext(context.Background(), opts.EntClient)

	factory.NewFactory(ctxObj)

	router.InstallRouters(opts.HertzServer)

	return opts.HertzServer.Run()
}
