package reactor

import (
	"context"
	"fmt"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/message"
	"github.com/kcloutie/event-reactor/pkg/template"
	"go.uber.org/zap"
)

var _ ReactorInterface = (*Reactor)(nil)

type Reactor struct {
	log *zap.Logger

	event       config.ReactorConfig
	Properties  []config.ReactorConfigProperty
	Name        string
	Description string
}

func NewTestReactor() *Reactor {

	return &Reactor{
		Name:        "testReactor",
		Description: "A test reactor that logs a message",
		Properties: []config.ReactorConfigProperty{
			{
				Name:        "message",
				Description: "The message to log. This field supports go templating",
				Required:    config.AsBoolPointer(true),
			},
		},
	}
}

func (v *Reactor) SetLogger(logger *zap.Logger) {
	v.log = logger
}
func (v *Reactor) GetName() string {
	return v.Name
}

func (v *Reactor) GetConfigExample() string {
	return ``
}

func (v *Reactor) GetDescription() string {
	return v.Description
}

func (v *Reactor) SetReactor(reactor config.ReactorConfig) {
	v.event = reactor
}

func (v *Reactor) ProcessEvent(ctx context.Context, data *message.EventData) error {
	logger := v.log.Sugar()
	_, err := HasRequiredProperties(v.event.Properties, v.GetRequiredPropertyNames())
	if err != nil {
		return err
	}
	message, err := v.event.Properties["message"].GetStringValue(ctx, v.log, data)
	if err != nil {
		return err
	}

	templateConfig := template.NewRenderTemplateOptions()
	SetGoTemplateOptionValues(ctx, v.log, &templateConfig, v.event.Properties)

	renderedMessage, err := template.RenderTemplateValues(ctx, message, fmt.Sprintf("%s_%s", data.ID, v.Name), data.AsMap(), []string{}, templateConfig)
	if err != nil {
		return err
	}

	logger.Info(string(renderedMessage))

	return nil
}

func (v *Reactor) GetHelp() string {
	return GetReactorHelp(v)
}

func (v *Reactor) GetProperties() []config.ReactorConfigProperty {
	return v.Properties

}

func (v *Reactor) GetRequiredPropertyNames() []string {
	return GetRequiredPropertyNames(v)
}

func ToPtrString(val string) *string {
	return &val
}
