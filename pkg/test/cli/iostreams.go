package cli

import (
	"bytes"
	"io"

	"github.com/kcloutie/event-reactor/pkg/cli"
	"github.com/kcloutie/event-reactor/pkg/params/settings"
)

func NewTestIOStreams() (*cli.IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	stream := &cli.IOStreams{
		In:     io.NopCloser(in),
		Out:    out,
		ErrOut: errOut,
	}
	settings.RootOptions.NoColor = true
	stream.SetColorEnabled(false)
	return stream, in, out, errOut
}
