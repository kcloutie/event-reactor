package webex

import (
	"context"
	"fmt"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/message"
	"github.com/kcloutie/event-reactor/pkg/params/settings"
	"github.com/kcloutie/event-reactor/pkg/reactor"
	"github.com/kcloutie/event-reactor/pkg/template"
	"github.com/kcloutie/event-reactor/pkg/webex"

	"go.uber.org/zap"
)

var _ reactor.ReactorInterface = (*Reactor)(nil)

type Reactor struct {
	Log           *zap.Logger
	reactorName   string
	reactorConfig config.ReactorConfig
}

type ReactorConfig struct {
	WebexCfg webex.WebexConfiguration
}

func New() *Reactor {
	return &Reactor{
		reactorName: "webex",
	}
}

func (v *Reactor) SetLogger(logger *zap.Logger) {
	v.Log = logger
}
func (v *Reactor) GetName() string {
	return v.reactorName
}

func (v *Reactor) GetDescription() string {
	return "This reactor will send a webex message. The message and card support go templating."
}

func (v *Reactor) GetConfigExample() string {
	return ``
}

func (v *Reactor) SetReactor(reactor config.ReactorConfig) {
	v.reactorConfig = reactor
}

func (v *Reactor) ProcessEvent(ctx context.Context, data *message.EventData) error {
	v.Log = v.Log.With(zap.String("reactor", v.reactorName))
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

	if reactorConfig.WebexCfg.Card != "" {
		err = reactorConfig.WebexCfg.SendWithCard()
	} else {
		err = reactorConfig.WebexCfg.SendMessage()
	}
	if err != nil {
		return err
	}
	return nil
}

func (v *Reactor) GetReactorConfig(ctx context.Context, data *message.EventData, log *zap.Logger) (*ReactorConfig, error) {

	templateConfig := template.NewRenderTemplateOptions()
	reactor.SetGoTemplateOptionValues(ctx, v.Log, &templateConfig, v.reactorConfig.Properties)

	message, err := v.reactorConfig.Properties["message"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	renderedMessage := []byte("empty message")
	if message != "" {
		renderedMessage, err = template.RenderTemplateValues(ctx, message, fmt.Sprintf("%s_%s/message", data.ID, v.reactorName), data.AsMap(), []string{}, templateConfig)
		if err != nil {
			return nil, err
		}
	}

	card, err := v.reactorConfig.Properties["card"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	renderedCard := []byte("")
	if card != "" {
		renderedCard, err = template.RenderTemplateValues(ctx, card, fmt.Sprintf("%s_%s/card", data.ID, v.reactorName), data.AsMap(), []string{}, templateConfig)
		if err != nil {
			return nil, err
		}
	}

	apiUrl, err := v.reactorConfig.Properties["apiUrl"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		apiUrl = ""
	}
	if apiUrl == "" {
		apiUrl = settings.WebexApiUrlDefault
	}

	token, err := v.reactorConfig.Properties["token"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if token == "" {
		return nil, fmt.Errorf("the token property was not supplied or was empty")
	}

	spaceId, err := v.reactorConfig.Properties["spaceId"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if spaceId == "" {
		return nil, fmt.Errorf("the spaceId property was not supplied or was empty")
	}

	config := &ReactorConfig{
		WebexCfg: webex.WebexConfiguration{
			Log:      log,
			ApiUrl:   apiUrl,
			ApiToken: token,
			SpaceId:  spaceId,
			Message:  string(renderedMessage),
			Card:     string(renderedCard),
		},
	}

	return config, nil
}

func (v *Reactor) GetProperties() []config.ReactorConfigProperty {
	return []config.ReactorConfigProperty{
		{
			Name:        "apiUrl",
			Description: fmt.Sprintf("The 'apiUrl' property is a required string that specifies the API endpoint for the Webex service. This is the base URL that the reactor will use to interact with the Webex API. Default: %s", settings.WebexApiUrlDefault),
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "token",
			Description: "The webex token to use for authentication. This token should have the necessary permissions to write messages to the Webex space.",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "spaceId",
			Description: "The spaceId to send the message to.",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "message",
			Description: "The message to send to the spaceId.",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "cardJson",
			Description: "The card to send to the spaceId. This should be a JSON string. The card will be sent as an attachment to the message. The card should be in the format of a Webex card. See https://developer.webex.com/docs/api/guides/cards for more information.",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
	}
}

func (v *Reactor) GetRequiredPropertyNames() []string {
	return reactor.GetRequiredPropertyNames(v)
}

func (v *Reactor) GetHelp() string {
	return reactor.GetReactorHelp(v)
}
