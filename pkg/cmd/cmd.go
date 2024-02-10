package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kcloutie/event-reactor/pkg/cli"
	"github.com/kcloutie/event-reactor/pkg/logger"
	"github.com/kcloutie/event-reactor/pkg/params/settings"
	"go.uber.org/zap"
)

const (
	NoColorFlag = "no-color"
)

func WriteCmdErrorToScreen(errorMessage string, ioStreams *cli.IOStreams, printErr bool, shouldExit bool) {
	if printErr {
		fmt.Fprintf(ioStreams.ErrOut, "%s %s\n", ioStreams.ColorScheme().FailureIcon(), ioStreams.ColorScheme().RedBold(errorMessage))
	}
	if shouldExit {
		os.Exit(1)
	}
}

func WriteCmdWarningToScreen(warningMessage string, ioStreams *cli.IOStreams, printErr bool, linePrefix string) {
	if printErr {
		PrintMessageToConsole(ioStreams.Out, fmt.Sprintf("%s%s%s %s\n", linePrefix, ioStreams.ColorScheme().Yellow("WARNING: "), ioStreams.ColorScheme().WarningIcon(), ioStreams.ColorScheme().Yellow(warningMessage)))
	}
}

func PrintMessageToConsole(writer io.Writer, message string) {
	if !settings.IsQuiet {
		fmt.Fprint(writer, message)
	}
}

func InitContextWithLogger(rootCmd string, subCmd string) context.Context {
	ctx := context.Background()
	return logger.WithCtx(ctx, logger.FromCtx(ctx).With(zap.String(logger.RootCommandKey, rootCmd)).With(zap.String(logger.SubCommandKey, subCmd)))
}

func CheckForUnknownArgsExitWhenFound(args []string, ioStreams *cli.IOStreams) {
	if len(args) > 0 {
		WriteCmdErrorToScreen(fmt.Sprintf("unknown arguments specified: '%s'", strings.Join(args, ", ")), ioStreams, true, true)
	}

}

func VerifyOutputParameterValue(output string) error {
	if output != "json" && output != "yaml" && output != "" {
		return fmt.Errorf("invalid output type '%s'", output)
	}
	// If someone wants the output to be in json or yaml, we dont want
	// to print anything to the screen that could interfere
	if output != "" {
		settings.IsQuiet = true
	}
	return nil
}
