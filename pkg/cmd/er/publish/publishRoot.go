package publish

import (
	"github.com/kcloutie/event-reactor/pkg/cli"
	"github.com/kcloutie/event-reactor/pkg/params"
	"github.com/spf13/cobra"
)

func Root(cliParams *params.Run, ioStreams *cli.IOStreams) *cobra.Command {
	cCmd := &cobra.Command{
		Use:     "publish",
		Aliases: []string{},
		Short:   "Publishes events",
	}
	cCmd.AddCommand(PubsubCommand(cliParams, ioStreams))
	return cCmd
}
