package get

import (
	"github.com/kcloutie/event-reactor/pkg/cli"
	"github.com/kcloutie/event-reactor/pkg/params"
	"github.com/spf13/cobra"
)

func Root(cliParams *params.Run, ioStreams *cli.IOStreams) *cobra.Command {
	cCmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{},
		Short:   "Gets data about the event reactor",
	}
	cCmd.AddCommand(ReactorCommand(cliParams, ioStreams))
	return cCmd
}
