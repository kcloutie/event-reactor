package cli

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"testing"

	"github.com/spf13/cobra"
)

// ExecuteCommand executes the root command passing the args and returns
// the output as a string and error
func ExecuteCommand(root *cobra.Command, args ...string) (string, error) {
	_, output, err := ExecuteCommandC(root, args...)
	return output, err
}

// ExecuteCommandC executes the root command passing the args and returns
// the root command, output as a string and error if any
func ExecuteCommandC(c *cobra.Command, args ...string) (*cobra.Command, string, error) {
	buf := new(bytes.Buffer)
	c.SetOutput(buf)
	c.SetArgs(args)
	c.SilenceUsage = true

	root, err := c.ExecuteC()

	return root, buf.String(), err
}

func TestCommand(t *testing.T, cCmd *cobra.Command, outBuf *bytes.Buffer, errOutBuf *bytes.Buffer, wantOut string, wantErrOut string, wantErr bool, args []string) {
	cCmd.SetArgs(args)
	err := cCmd.Execute()
	if (err != nil) != wantErr {
		t.Errorf("%s() error = %v, wantErr %v", cCmd.Use, err, wantErr)
	}

	TestOutput(t, cCmd.Use, outBuf, errOutBuf, wantOut, wantErrOut, wantErr)
}

func TestOutput(t *testing.T, commandName string, outBuf *bytes.Buffer, errOutBuf *bytes.Buffer, wantOut string, wantErrOut string, wantErr bool) {

	rOut, err := regexp.Compile(wantOut)
	if err != nil {
		t.Errorf("%s() failed to compile wantOut - %v", commandName, err)
		return
	}

	out, err := io.ReadAll(outBuf)
	if err != nil {
		t.Errorf("%s() unable to read output - %v", commandName, err)
		return
	}
	ttt := rOut.MatchString(wantOut)
	fmt.Println(ttt)
	if !rOut.Match(out) {
		t.Errorf("%s() out was:\n'%v'\n wanted:\n'%v'\n", commandName, string(out), wantOut)
	}

	rErrOut, err := regexp.Compile(wantErrOut)
	if err != nil {
		t.Errorf("%s() failed to compile wantErrOut - %v", commandName, err)
		return
	}

	outErr, err := io.ReadAll(errOutBuf)
	if err != nil {
		t.Errorf("%s() unable to read error output - %v", commandName, err)
		return
	}

	if wantErrOut == "" {
		if string(outErr) != "" {
			t.Errorf("%s() outErr was:\n%v\n wanted:\n%v\n", commandName, string(outErr), wantErrOut)
			return
		}
		return
	}

	if !rErrOut.Match(outErr) {
		t.Errorf("%s() outErr was:\n%v\n wanted:\n%v\n", commandName, string(outErr), wantErrOut)
	}

}
