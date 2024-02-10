package main

import (
	"os"

	"github.com/kcloutie/event-reactor/pkg/cmd/er"
	"github.com/kcloutie/event-reactor/pkg/params"
)

func main() {
	cliParams := params.New()
	cli := er.Root(cliParams)

	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
