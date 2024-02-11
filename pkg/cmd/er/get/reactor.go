package get

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/kcloutie/event-reactor/pkg/adapter"
	"github.com/kcloutie/event-reactor/pkg/cli"
	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/logger"
	"github.com/kcloutie/event-reactor/pkg/reactor"
	"github.com/spf13/cobra"

	"github.com/kcloutie/event-reactor/pkg/cmd"
	"github.com/kcloutie/event-reactor/pkg/params"
)

type ReactorCmdOptions struct {
	IoStreams *cli.IOStreams
	CliOpts   *cli.CliOpts
	Name      string
}

func ReactorCommand(run *params.Run, rootOpts *RootCmdOption, ioStreams *cli.IOStreams) *cobra.Command {
	options := &ReactorCmdOptions{}
	cCmd := &cobra.Command{
		Use:     "reactor",
		Aliases: []string{},
		Short:   "Displays all reactors, or details a specific reactor if a name is provided.",
		Long: heredoc.Docf(`
		Enumerates all reactors, providing names and descriptions, or details a specific reactor when a name is provided.
		`, "`"),
		Example: heredoc.Doc(`
			# List all reactors
			er get reactor 

			# Show details of the powershell reactor
			er get reactor --name powershell
		`),
		Run: func(cCmd *cobra.Command, args []string) {
			ctx := cmd.InitContextWithLogger("get", "reactor")

			options.IoStreams = ioStreams
			options.CliOpts = cli.NewCliOptions()
			options.IoStreams.SetColorEnabled(!rootOpts.NoColorFlag)
			cmd.CheckForUnknownArgsExitWhenFound(args, ioStreams)

			rf := adapter.GetReactorNewFunctions(false)

			log := logger.FromCtx(ctx)
			longestName := 0
			reactors := map[string]reactor.ReactorInterface{}
			for name, rFunc := range rf {
				reactors[name] = rFunc(log, config.ReactorConfig{})
				if len(name) > longestName {
					longestName = len(name)
				}
			}
			if options.Name != "" {
				r, ok := reactors[options.Name]
				if !ok {
					cmd.PrintMessageToConsole(ioStreams.Out, fmt.Sprintf("Reactor with name %s not found\n", options.Name))
					return
				}
				cmd.PrintMessageToConsole(ioStreams.Out, fmt.Sprintf("%s\n", ioStreams.ColorScheme().BlueBold(r.GetHelp())))
				return
			}
			cmd.PrintMessageToConsole(ioStreams.Out, "\n")
			for name, r := range reactors {
				// spacingsLen := longestName - (len(name) - 2)
				// if spacingsLen < 0 {
				// 	spacingsLen = 1
				// }
				// spacing := strings.Repeat(" ", spacingsLen)
				cmd.PrintMessageToConsole(ioStreams.Out, fmt.Sprintf("%s: %s\n\n", ioStreams.ColorScheme().BlueBold(strings.ToUpper(name)), r.GetDescription()))
			}

		},
	}
	cCmd.Flags().StringVarP(&options.Name, "name", "n", "", "The name of the reactor to display details for. If not provided, all reactors are displayed.")
	return cCmd
}
