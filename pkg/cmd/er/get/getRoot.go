package get

import (
	"github.com/kcloutie/event-reactor/pkg/cli"
	"github.com/kcloutie/event-reactor/pkg/cmd"
	"github.com/kcloutie/event-reactor/pkg/params"
	"github.com/spf13/cobra"
)

type RootCmdOption struct {
	NoColorFlag bool
	Output      string
	WorkingDir  string
}

func Root(cliParams *params.Run, ioStreams *cli.IOStreams) *cobra.Command {
	cCmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{},
		Short:   "Gets data about the event reactor",
	}

	rootOpts := &RootCmdOption{}

	cCmd.PersistentFlags().BoolVarP(&rootOpts.NoColorFlag, cmd.NoColorFlag, "C", false, "Disable coloring")
	cCmd.PersistentFlags().StringVarP(&rootOpts.Output, "output", "o", "", "Output format. One of: (json, yaml)")
	cCmd.PersistentFlags().StringVarP(&rootOpts.WorkingDir, "cwd", "w", "", "Current working directory")

	cCmd.AddCommand(ReactorCommand(cliParams, rootOpts, ioStreams))
	cCmd.AddCommand(GcpSecretCommand(cliParams, rootOpts, ioStreams))
	return cCmd
}
