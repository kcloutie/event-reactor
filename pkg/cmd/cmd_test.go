package cmd

import (
	"testing"

	testcli "github.com/kcloutie/event-reactor/pkg/test/cli"
)

func TestWriteCmdErrorToScreen(t *testing.T) {
	type args struct {
		errorMessage string
		printErr     bool
		shouldExit   bool
	}
	tests := []struct {
		name       string
		args       args
		wantErrOut string
	}{
		{
			name: "print error",
			args: args{
				errorMessage: "error",
				printErr:     true,
				shouldExit:   false,
			},
			wantErrOut: "X error",
		},
		{
			name: "dont print error",
			args: args{
				errorMessage: "error",
				printErr:     false,
				shouldExit:   false,
			},
			wantErrOut: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ioStreams, _, outBuf, errOutBuf := testcli.NewTestIOStreams()
			WriteCmdErrorToScreen(tt.args.errorMessage, ioStreams, tt.args.printErr, tt.args.shouldExit)
			testcli.TestOutput(t, "WriteCmdErrorToScreen", outBuf, errOutBuf, "", tt.wantErrOut, false)
		})
	}
}

func TestWriteCmdWarningToScreen(t *testing.T) {
	type args struct {
		warningMessage string
		printErr       bool
		linePrefix     string
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
	}{
		{
			name: "print warning",
			args: args{
				warningMessage: "warning",
				printErr:       true,
				linePrefix:     "",
			},
			wantOut: "WARNING: ! warning",
		},
		{
			name: "do not print warning",
			args: args{
				warningMessage: "warning",
				printErr:       false,
				linePrefix:     "",
			},
			wantOut: "",
		},
		{
			name: "print warning prefix",
			args: args{
				warningMessage: "warning",
				printErr:       true,
				linePrefix:     "--",
			},
			wantOut: "--WARNING: ! warning",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ioStreams, _, outBuf, errOutBuf := testcli.NewTestIOStreams()
			WriteCmdWarningToScreen(tt.args.warningMessage, ioStreams, tt.args.printErr, tt.args.linePrefix)
			testcli.TestOutput(t, "WriteCmdErrorToScreen", outBuf, errOutBuf, tt.wantOut, "", false)
		})
	}
}

func TestVerifyOutputParameterValue(t *testing.T) {
	type args struct {
		output string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "blank",
			args: args{
				output: "json",
			},
			wantErr: false,
		},
		{
			name: "json",
			args: args{
				output: "json",
			},
			wantErr: false,
		},
		{
			name: "yaml",
			args: args{
				output: "yaml",
			},
			wantErr: false,
		},
		{
			name: "fail unknown",
			args: args{
				output: "wrong",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := VerifyOutputParameterValue(tt.args.output); (err != nil) != tt.wantErr {
				t.Errorf("VerifyOutputParameterValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
