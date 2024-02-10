package publish

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/kcloutie/event-reactor/pkg/cli"
	"github.com/kcloutie/event-reactor/pkg/gcp"
	"github.com/kcloutie/event-reactor/pkg/params/settings"
	"github.com/spf13/cobra"

	"github.com/kcloutie/event-reactor/pkg/cmd"
	"github.com/kcloutie/event-reactor/pkg/params"
)

type ReactorCmdOptions struct {
	IoStreams  *cli.IOStreams
	CliOpts    *cli.CliOpts
	Data       string
	DataFile   string
	Topic      string
	Project    string
	Attributes []string
}

func PubsubCommand(run *params.Run, ioStreams *cli.IOStreams) *cobra.Command {
	options := &ReactorCmdOptions{}
	cCmd := &cobra.Command{
		Use:     "pubsub",
		Aliases: []string{"ps"},
		Short:   "Publishes messages to a pubsub topic",
		Long: heredoc.Docf(`
		 Publishes message to a pubsub topic.
		`, "`"),
		Example: heredoc.Doc(`
			# List all reactors
			er publish pubsub --data '{"name": "John"}' --topic "my-topic" --project "my-project"

			# Show details of the powershell reactor
			er get reactor --name powershell
		`),
		Run: func(cCmd *cobra.Command, args []string) {
			ctx := cmd.InitContextWithLogger("get", "reactor")

			options.IoStreams = ioStreams
			options.CliOpts = cli.NewCliOptions()
			options.IoStreams.SetColorEnabled(!settings.RootOptions.NoColor)
			cmd.CheckForUnknownArgsExitWhenFound(args, ioStreams)

			options.PublishMessage(ctx)

		},
	}
	cCmd.Flags().StringVarP(&options.Data, "data", "d", "", "The content of the message to publish")
	cCmd.Flags().StringVarP(&options.DataFile, "data-file", "f", "", "path to a file containing the message to publish")
	cCmd.Flags().StringVarP(&options.Topic, "topic", "t", "", "The topic to publish the message to")
	cCmd.Flags().StringVarP(&options.Project, "project", "p", "", "The gcp project the topic is in")
	cCmd.Flags().StringArrayVarP(&options.Attributes, "attribute", "a", []string{}, "The attributes to publish with the message")
	return cCmd
}

func (o *ReactorCmdOptions) PublishMessage(ctx context.Context) {

	if o.DataFile != "" {
		fileBytes, err := os.ReadFile(o.DataFile)
		if err != nil {
			cmd.WriteCmdErrorToScreen(fmt.Sprintf("failed to read the data file - %v", err), o.IoStreams, true, true)
		}
		o.Data = string(fileBytes)
	}
	if o.Data == "" {
		cmd.WriteCmdErrorToScreen("data or data-file must be provided and must contain data", o.IoStreams, true, true)
	}
	if o.Topic == "" {
		cmd.WriteCmdErrorToScreen("topic must be provided", o.IoStreams, true, true)
	}
	if o.Project == "" {
		cmd.WriteCmdErrorToScreen("project must be provided", o.IoStreams, true, true)
	}
	attrs := map[string]string{}
	for _, attr := range o.Attributes {
		parts := strings.Split(attr, "=")
		if len(parts) != 2 {
			cmd.WriteCmdErrorToScreen(fmt.Sprintf("invalid attribute format - '%v'. Must be in the form of <key>=<value>", attr), o.IoStreams, true, true)
		}
		attrs[parts[0]] = parts[1]
	}
	id, err := gcp.PublishEvent(o.Project, o.Topic, []byte(o.Data), attrs)
	if err != nil {
		cmd.WriteCmdErrorToScreen(fmt.Sprintf("failed to publish the message - %v", err), o.IoStreams, true, true)
	}
	fmt.Fprintf(o.IoStreams.Out, "Published message to topic '%v' with id '%v'\n", o.Topic, id)
}
