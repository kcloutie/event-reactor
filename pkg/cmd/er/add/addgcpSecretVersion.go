package add

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
	IoStreams      *cli.IOStreams
	CliOpts        *cli.CliOpts
	Project        string
	Name           string
	SecretData     string
	SecretDataPath string
}

func AddGcpSecretVersionCommand(run *params.Run, rootOpts *RootCmdOption, ioStreams *cli.IOStreams) *cobra.Command {
	options := &GcpSecretOptions{}
	cCmd := &cobra.Command{
		Use:     "gcp-secret-version",
		Aliases: []string{"version", "gcpsecretversion"},
		Short:   "Adds a new version to a GCP secret",
		Long: heredoc.Docf(`
			Add a GCP secret version.

			Required flags:

			- %[1]s--project%[1]s
			- %[1]s--name%[1]s
		`, "`"),
		Example: heredoc.Doc(`
			# log into GCP
			gcloud auth login --update-adc

			# get secret from GCP secret manager
			cldctl add gcp-secret-version -p gcp-proj -n some-secret -d "new secret version"
		`),
		RunE: func(cCmd *cobra.Command, args []string) error {

			ctx := cmd.InitContextWithLogger("add", "gcp-secret-version")

			options.IoStreams = ioStreams
			options.CliOpts = cli.NewCliOptions()
			options.IoStreams.SetColorEnabled(!rootOpts.NoColorFlag)

			cmd.CheckForUnknownArgsExitWhenFound(args, ioStreams)

			version, err := gcp.AddVersion(ctx, nil, options.Project, options.Name, options.SecretData)
			if err != nil {
				cmd.WriteCmdErrorToScreen(fmt.Sprintf("failed to get the secret: %v", err), options.IoStreams, true, true)
			}

			fmt.Fprintf(ioStreams.Out, "%s\n", ioStreams.ColorScheme().Green(fmt.Sprintf("Version \"%s\" added", version)))

			return nil
		},
	}

	cCmd.PersistentFlags().StringVarP(&options.Project, "project", "p", "", "GCP project where the secret exists")
	cCmd.PersistentFlags().StringVarP(&options.Name, "name", "n", "", "Name of the secret")
	cCmd.PersistentFlags().StringVarP(&options.SecretData, "secret-data", "d", "", "The contents of the new secret version")
	cCmd.PersistentFlags().StringVar(&options.SecretDataPath, "secret-data-path", "", "The path to the file containing the new secret version content")

	return cCmd
}
