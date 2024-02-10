package version

import (
	"testing"

	v "github.com/kcloutie/event-reactor/pkg/params/version"
	testcli "github.com/kcloutie/event-reactor/pkg/test/cli"
)

func TestVersionCommandExecute(t *testing.T) {
	v.BuildTime = "buildTime"
	v.BuildVersion = "version"
	v.Commit = "commit"
	tests := []struct {
		name        string
		commandName string
		args        []string
		wantErr     bool
		wantErrOut  string
		wantOut     string
	}{
		{
			name:        "basic",
			commandName: "Version",
			args:        []string{},
			wantOut: `Version:        version
Commit:         commit
Build Time:     buildTime`,
		},
		{
			name:        "json",
			commandName: "Version",
			args:        []string{"-o", "json"},
			wantOut:     `{"buildTime":"buildTime","commit":"commit","version":"version"}`,
		},
		{
			name:        "yaml",
			commandName: "Version",
			args:        []string{"-o", "yaml"},
			wantOut: `buildTime: buildTime
commit: commit
version: version`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ioStreams, _, outBuf, errOutBuf := testcli.NewTestIOStreams()
			cCmd := VersionCommand(ioStreams)
			testcli.TestCommand(t, cCmd, outBuf, errOutBuf, tt.wantOut, tt.wantErrOut, tt.wantErr, tt.args)
			// cCmd.SetArgs(tt.args)
			// err := cCmd.Execute()
			// if (err != nil) != tt.wantErr {
			// 	t.Errorf("%s() error = %v, wantErr %v", tt.commandName, err, tt.wantErr)
			// }

			// rOut, err := regexp.Compile(tt.wantOut)
			// if err != nil {
			// 	t.Errorf("%s() failed to compile wantOut - %v", tt.commandName, err)
			// 	return
			// }

			// out, err := io.ReadAll(outBuf)
			// if err != nil {
			// 	t.Errorf("%s() unable to read output - %v", tt.commandName, err)
			// 	return
			// }
			// ttt := rOut.MatchString(tt.wantOut)
			// fmt.Println(ttt)
			// if !rOut.Match(out) {
			// 	t.Errorf("%s() out was:\n'%v'\n wanted:\n'%v'\n", tt.commandName, string(out), tt.wantOut)
			// }

			// rErrOut, err := regexp.Compile(tt.wantErrOut)
			// if err != nil {
			// 	t.Errorf("%s() failed to compile wantErrOut - %v", tt.commandName, err)
			// 	return
			// }

			// outErr, err := io.ReadAll(errOutBuf)
			// if err != nil {
			// 	t.Errorf("%s() unable to read error output - %v", tt.commandName, err)
			// 	return
			// }

			// if tt.wantErrOut == "" {
			// 	if string(outErr) != "" {
			// 		t.Errorf("%s() outErr was:\n%v\n wanted:\n%v\n", tt.commandName, string(outErr), tt.wantErrOut)
			// 		return
			// 	}
			// 	return
			// }

			// if rErrOut.Match(outErr) {
			// 	t.Errorf("%s() outErr was:\n%v\n wanted:\n%v\n", tt.commandName, string(outErr), tt.wantErrOut)
			// }

		})
	}
}
