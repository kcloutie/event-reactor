package cli

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

type CliOpts struct {
	NoColoring bool
	AskOpts    survey.AskOpt
}

func NewAskOpts(opt *survey.AskOptions) error {
	opt.Stdio = terminal.Stdio{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
	return nil
}

func NewCliOptions() *CliOpts {
	return &CliOpts{
		AskOpts: NewAskOpts,
	}
}
