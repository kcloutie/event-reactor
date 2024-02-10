package params

import (
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

type CliOpts struct {
	NoColoring bool
	AskOpts    survey.AskOpt
}

func NewCliOptions() *CliOpts {
	return &CliOpts{
		AskOpts: func(opt *survey.AskOptions) error {
			opt.Stdio = terminal.Stdio{
				In:  os.Stdin,
				Out: os.Stdout,
				Err: os.Stderr,
			}
			return nil
		},
	}
}

func (c *CliOpts) Ask(qss []*survey.Question, ans interface{}) error {
	return survey.Ask(qss, ans, c.AskOpts)
}
