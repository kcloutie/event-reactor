package powershell

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/message"
	"github.com/kcloutie/event-reactor/pkg/pwsh"
	"github.com/kcloutie/event-reactor/pkg/reactor"
	"github.com/kcloutie/event-reactor/pkg/template"

	"go.uber.org/zap"
)

var _ reactor.ReactorInterface = (*Reactor)(nil)

type Reactor struct {
	Log           *zap.Logger
	reactorName   string
	reactorConfig config.ReactorConfig
}

type ReactorConfig struct {
	PwshConfig pwsh.PwshExecConfig
	Command    string
	Parameters map[string]interface{}
}

func New() *Reactor {
	return &Reactor{
		reactorName: "powershell",
	}
}

func (v *Reactor) SetLogger(logger *zap.Logger) {
	v.Log = logger
}
func (v *Reactor) GetName() string {
	return v.reactorName
}

func (v *Reactor) GetConfigExample() string {
	return `
  reactorConfigs:
  - name: test_powershell
    celExpressionFilter: attributes.test == 'powershell'
    disabled: false
    type: powershell
    properties:
      command:
        value: Get-Date
      parameters:
        value:
          Format: yyyy-MM-dd HH:mm:ss
          AsUTC:
`
}

func (v *Reactor) GetDescription() string {
	return "This reactor executes a powershell command. The parameter values support go templating."
}

func (v *Reactor) SetReactor(reactor config.ReactorConfig) {
	v.reactorConfig = reactor
}

func (v *Reactor) ProcessEvent(ctx context.Context, data *message.EventData) error {

	_, err := reactor.HasRequiredProperties(v.reactorConfig.Properties, v.GetRequiredPropertyNames())
	if err != nil {
		return err
	}

	templateConfig := template.NewRenderTemplateOptions()
	reactor.SetGoTemplateOptionValues(ctx, v.Log, &templateConfig, v.reactorConfig.Properties)

	reactorConfig, err := v.GetReactorConfig(ctx, data, v.Log)
	if err != nil {
		return err
	}
	v.Log = v.Log.With(zap.String("reactor", v.reactorName)).With(zap.String("command", reactorConfig.Command))
	jsonResults, err := reactorConfig.PwshConfig.ExecuteWithParamsFile(ctx, reactorConfig.Command, reactorConfig.Parameters)
	v.Log.Debug("Powershell command executed", zap.String("results", string(jsonResults)))
	if err != nil {
		return err
	}

	return nil
}

func (v *Reactor) GetReactorConfig(ctx context.Context, data *message.EventData, log *zap.Logger) (*ReactorConfig, error) {
	config := ReactorConfig{
		PwshConfig: pwsh.PwshExecConfig{
			NoColor: true,
		},
	}

	templateConfig := template.NewRenderTemplateOptions()
	reactor.SetGoTemplateOptionValues(ctx, v.Log, &templateConfig, v.reactorConfig.Properties)

	// ===================================================================================
	// Get jsonDepth
	// ===================================================================================
	jsonDepth := 4
	jsonDepthStr, err := v.reactorConfig.Properties["jsonDepth"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if jsonDepthStr != "" {
		jsonDepth, err = strconv.Atoi(jsonDepthStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied parameters '%v' to an integer. Error: %v", jsonDepthStr, err)
		}
	}
	config.PwshConfig.Depth = jsonDepth

	// ===================================================================================
	// Get command
	// ===================================================================================
	command, err := v.reactorConfig.Properties["command"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if command == "" {
		return nil, fmt.Errorf("the command property was not supplied or was empty")
	}
	config.Command = command

	// ===================================================================================
	// Get parameters
	// ===================================================================================

	parameters, err := v.reactorConfig.Properties["parameters"].GetMapStringInterfaceValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}

	for k, propv := range parameters {
		switch val := propv.(type) {
		case string:
			renderedSubject, err := template.RenderTemplateValues(ctx, val, fmt.Sprintf("%s_%s/param_%s", data.ID, v.reactorName, k), data.AsMap(), []string{}, templateConfig)
			if err != nil {
				return nil, err
			}
			parameters[k] = string(renderedSubject)
		}
	}
	config.Parameters = parameters

	return &config, nil
}

func (v *Reactor) GetHelp() string {
	return reactor.GetReactorHelp(v)
}

func (v *Reactor) GetProperties() []config.ReactorConfigProperty {
	return []config.ReactorConfigProperty{
		{
			Name:        "command",
			Description: "The powershell command to execute",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "parameters",
			Description: "The parameters to pass to the powershell command. This field supports go templating",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeMapString,
		},
		{
			Name:        "jsonDepth",
			Description: "The maximum depth the JSON is allowed to have. Defaults to 4",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
	}
}

func (v *Reactor) GetRequiredPropertyNames() []string {
	return reactor.GetRequiredPropertyNames(v)
}
