package run

import (
	"github.com/kcloutie/event-reactor/pkg/cli"
	"github.com/kcloutie/event-reactor/pkg/params"
	"github.com/spf13/cobra"
)

func Root(cliParams *params.Run, ioStreams *cli.IOStreams) *cobra.Command {
	cCmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{},
		Short:   "Runs the web/api server",
	}
	cCmd.AddCommand(ServerCommand(cliParams, ioStreams))
	return cCmd
}
