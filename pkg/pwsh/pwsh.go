package pwsh

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kcloutie/event-reactor/pkg/encoding"
)

// https://stackoverflow.com/questions/65331558/how-to-call-powershell-from-go-faster
// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html
const (
	PowerShellExe = "pwsh"
)

type PwshExecConfig struct {
	Depth   int
	NoColor bool
}

func (o *PwshExecConfig) ExecuteRaw(ctx context.Context, command string) ([]byte, error) {
	encodedCommand := encoding.EncodeStringToUtf16(command)
	out, err := exec.Command(PowerShellExe, "-o", "Text", "-nologo", "-noprofile", "-NonInteractive", "-EncodedCommand", encodedCommand).CombinedOutput()
	if err != nil {
		return out, err
	}
	return out, nil
}

func (o *PwshExecConfig) Execute(ctx context.Context, command string) ([]byte, error) {
	out, err := exec.Command(PowerShellExe, "-nologo", "-noprofile", "-NonInteractive", "-Command", command, "|", "ConvertTo-Json", "-Depth", "2", "-Compress").CombinedOutput()
	if err != nil {
		return out, err
	}
	return out, nil
}

func (o *PwshExecConfig) ExecuteWithParamsFile(ctx context.Context, command string, parameters map[string]interface{}) ([]byte, error) {
	fullCommand, err := WrapCommand(command, parameters, o.Depth, o.NoColor)
	if err != nil {
		return []byte{}, err
	}

	td, err := os.MkdirTemp("", "pwsh")
	if err != nil {
		return []byte{}, err
	}

	scriptPath := filepath.Join(td, "script.ps1")
	defer os.RemoveAll(td)
	// fmt.Printf("Script File: %s\n", scriptPath)
	err = os.WriteFile(scriptPath, []byte(fullCommand), 0644)
	if err != nil {
		return []byte{}, err
	}
	cmdResults := exec.Command(PowerShellExe, "-o", "Text", "-nologo", "-noprofile", "-NonInteractive", "-File", scriptPath)

	out, err := cmdResults.CombinedOutput()

	if err != nil {
		return out, err
	}

	if cmdResults.ProcessState.ExitCode() != 0 {
		return out, fmt.Errorf(string(out))
	}

	return out, nil
}

func NewPwshExecConfig() PwshExecConfig {
	return PwshExecConfig{
		Depth:   2,
		NoColor: false,
	}
}

func WrapCommand(command string, parameters map[string]interface{}, depth int, noColor bool) (string, error) {

	funcParams, err := json.Marshal(parameters)
	if err != nil {
		return "", err
	}

	term := ""
	exceptionMsg := ""
	if noColor {
		term = "$env:TERM='dumb';"
		exceptionMsg = ".Exception.Message"
	}

	//  *>&1
	sb := fmt.Sprintf(`
	Trap {
		$_%s
		Write-Output 'AN ERROR HAS OCCURRED!!'
		Exit 1
	}

	$funcParams = @'
	%s
'@ | ConvertFrom-Json -Depth %v -AsHashtable -WarningAction Stop
	$WarningPreference='Stop'
	$results = %s %s @funcParams
	$p = @{
		Depth         = %v
		Compress      = $true
		WarningAction = "Stop"
	}
	if ($results -is [array]) {
		$p.add('AsArray', $true)
	}
	$results | ConvertTo-Json @p
	`,
		exceptionMsg, string(funcParams), depth, term, command, depth)

	return sb, nil
}
