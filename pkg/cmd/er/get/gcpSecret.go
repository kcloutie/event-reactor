package get

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/kcloutie/event-reactor/pkg/cli"
	"github.com/kcloutie/event-reactor/pkg/cmd"
	"github.com/kcloutie/event-reactor/pkg/gcp"
	"github.com/kcloutie/event-reactor/pkg/params"
	"github.com/spf13/cobra"
)

type GcpSecretOptions struct {
	IoStreams *cli.IOStreams
	CliOpts   *cli.CliOpts
	Project   string
	Name      string
	Version   string
}

func GcpSecretCommand(run *params.Run, rootOpts *RootCmdOption, ioStreams *cli.IOStreams) *cobra.Command {
	options := &GcpSecretOptions{}
	cCmd := &cobra.Command{
		Use:     "gcp-secret",
		Aliases: []string{"secret", "gcpsecret"},
		Short:   "Get a GCP secret value",
		Long: heredoc.Docf(`
			Get a GCP secret value for a given project.

			Required flags:

			- %[1]s--project%[1]s
			- %[1]s--name%[1]s
		`, "`"),
		Example: heredoc.Doc(`
			# log into GCP
			gcloud auth login --update-adc

			# get secret from GCP secret manager
			cldctl get gcp-secret -p gcp-proj -n some-secret
		`),
		RunE: func(cCmd *cobra.Command, args []string) error {

			ctx := cmd.InitContextWithLogger("get", "gcp-secret")

			options.IoStreams = ioStreams
			options.CliOpts = cli.NewCliOptions()
			options.IoStreams.SetColorEnabled(!rootOpts.NoColorFlag)

			cmd.CheckForUnknownArgsExitWhenFound(args, ioStreams)

			secret, err := gcp.GetSecret(ctx, nil, options.Project, options.Name, options.Version)
			if err != nil {
				cmd.WriteCmdErrorToScreen(fmt.Sprintf("failed to get the secret: %v", err), options.IoStreams, true, true)
			}

			fmt.Fprintf(ioStreams.Out, "%s", ioStreams.ColorScheme().Gray(secret))

			return nil
		},
	}

	cCmd.PersistentFlags().StringVarP(&options.Project, "project", "p", "", "GCP project where the secret exists")
	cCmd.PersistentFlags().StringVarP(&options.Name, "name", "n", "", "Name of the secret")
	cCmd.PersistentFlags().StringVarP(&options.Version, "version", "v", "latest", "Version of the secret")

	return cCmd
}
