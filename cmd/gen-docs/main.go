package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kcloutie/event-reactor/pkg/cmd/er"
	"github.com/kcloutie/event-reactor/pkg/doc"
	"github.com/kcloutie/event-reactor/pkg/params"
	"github.com/spf13/pflag"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	standardDoc := flags.BoolP("standard", "", false, "Generate standard docs")
	customDoc := flags.BoolP("custom", "", false, "Generate custom docs")
	dir := flags.StringP("doc-path", "", "", "Path directory where you want generate doc files")
	help := flags.BoolP("help", "h", false, "Help about any command")

	if err := flags.Parse(args); err != nil {
		return err
	}

	if *help {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n%s", filepath.Base(args[0]), flags.FlagUsages())
		return nil
	}

	if *dir == "" {
		return fmt.Errorf("error: --doc-path not set")
	}

	cliParams := params.New()
	cli := er.Root(cliParams)

	if err := os.MkdirAll(*dir, 0755); err != nil {
		return err
	}

	if *standardDoc {
		if err := doc.GenMarkdownTree(cli, *dir); err != nil {
			return err
		}
	}

	if *customDoc {
		if err := doc.GenMarkdownTreeCustom(cli, *dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	return nil
}

func filePrepender(filename string) string {
	return ""
}

func linkHandler(name string) string {
	return name
}
