package er

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/kcloutie/event-reactor/pkg/cli"
	"github.com/kcloutie/event-reactor/pkg/cmd"
	"github.com/kcloutie/event-reactor/pkg/cmd/er/add"
	"github.com/kcloutie/event-reactor/pkg/cmd/er/get"
	"github.com/kcloutie/event-reactor/pkg/cmd/er/publish"
	"github.com/kcloutie/event-reactor/pkg/cmd/er/run"
	"github.com/kcloutie/event-reactor/pkg/cmd/er/version"
	"github.com/kcloutie/event-reactor/pkg/logger"
	"github.com/kcloutie/event-reactor/pkg/params"
	"github.com/kcloutie/event-reactor/pkg/params/settings"
	"github.com/spf13/cobra"
)

var (
	showVersion = false
	ioStreams   = cli.NewIOStreams()
)

func Root(cliParams *params.Run) *cobra.Command {
	cCmd := &cobra.Command{
		Use:   "er",
		Short: "er (Event-Reactor) is a cli/api tool for reacting to events",
		Long: heredoc.Doc(`
			er (Event-Reactor) is a cli/api tool for reacting to events
		`),
		SilenceUsage: false,
		PersistentPreRun: func(cCmd *cobra.Command, args []string) {
			lgr := logger.Get()
			lgr.Info("Starting application")
			if settings.DebugModeEnabled || os.Getenv(settings.DebugModeLoggerEnvVar) != "" {
				lgr.Info("Debugging has been enabled!")
			}

		},
		RunE: func(cCmd *cobra.Command, args []string) error {
			if showVersion {
				vopts := version.VersionCmdOptions{
					IoStreams: ioStreams,
					CliOpts:   cli.NewCliOptions(),
					Output:    "",
				}
				vopts.IoStreams.SetColorEnabled(!settings.RootOptions.NoColor)
				vopts.PrintVersion(context.Background())
				return nil
			}
			return fmt.Errorf("no command was specified")
		},
		Annotations: map[string]string{
			"commandType": "main",
		},
	}

	cCmd.PersistentFlags().BoolVar(&settings.DebugModeEnabled, "debug", false, "When set, additional output around debugging is output to the screen")
	cCmd.PersistentFlags().BoolVarP(&settings.RootOptions.NoColor, cmd.NoColorFlag, "C", false, "Disable coloring")
	cCmd.PersistentFlags().BoolVar(&showVersion, "version", false, "Show the version")
	cCmd.AddCommand(version.VersionCommand(ioStreams))
	cCmd.AddCommand(run.Root(cliParams, ioStreams))
	cCmd.AddCommand(get.Root(cliParams, ioStreams))
	cCmd.AddCommand(publish.Root(cliParams, ioStreams))
	cCmd.AddCommand(add.Root(cliParams, ioStreams))

	return cCmd
}

func Test_ExecuteCommand(cliParams *params.Run, args []string) (string, string, error) {
	cmd := Root(cliParams)
	b := bytes.NewBufferString("")
	be := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetErr(be)
	cmd.SetArgs(args)
	err := cmd.Execute()
	if err != nil {
		return "", "", err
	}

	out, err := io.ReadAll(b)
	if err != nil {
		return "", "", err
	}
	outErr, err := io.ReadAll(be)
	if err != nil {
		return "", "", err
	}

	return string(out), string(outErr), nil
}
