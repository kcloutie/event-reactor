package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/http"
	"github.com/kcloutie/event-reactor/pkg/message"
	"github.com/kcloutie/event-reactor/pkg/params/settings"
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
	WebhookConfig http.WebhookConfig
}

func New() *Reactor {
	return &Reactor{
		reactorName: "webhook",
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
  - name: test_webhook
  disabled: false
  type: webhook
  properties:
    url:
      value: http://localhost:8080/echo
    webhookSecret:
      value: test123
    bodyTemplate: 
      value: '{"prop1": "{{ .data.prop1 }}"}'
    bearerToken:
      value: "faketoken"
    signatureHeader:
      value: "X-Hub-Signature-256"
    additionalHeaders:
      value:
        X-My-Header: "{{ .data.prop1 }}"
`
}

func (v *Reactor) GetDescription() string {
	return "This reactor sends a webhook to a specified URL. The payload of the webhook is the event data."
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

	err = reactorConfig.WebhookConfig.SendWebhook()
	if err != nil {
		return err
	}

	return nil
}

func (v *Reactor) GetReactorConfig(ctx context.Context, data *message.EventData, log *zap.Logger) (*ReactorConfig, error) {

	config := &ReactorConfig{
		WebhookConfig: http.WebhookConfig{
			Log: v.Log,
		},
	}

	templateConfig := template.NewRenderTemplateOptions()
	reactor.SetGoTemplateOptionValues(ctx, v.Log, &templateConfig, v.reactorConfig.Properties)

	// ===================================================================================
	// Get url
	// ===================================================================================
	url, err := v.reactorConfig.Properties["url"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if url == "" {
		return nil, fmt.Errorf("the url property was not supplied or was empty")
	}
	config.WebhookConfig.Url = url

	// ===================================================================================
	// Get webhookSecret
	// ===================================================================================
	webhookSecret, err := v.reactorConfig.Properties["webhookSecret"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	config.WebhookConfig.HookSecret = webhookSecret

	// ===================================================================================
	// Get bodyTemplate
	// ===================================================================================
	bodyTemplate, err := v.reactorConfig.Properties["bodyTemplate"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	config.WebhookConfig.Body = bodyTemplate

	if config.WebhookConfig.Body == "" {
		payloadContent, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal the event data to json. Error: %v", err)
		}
		config.WebhookConfig.Body = string(payloadContent)
	} else {
		if strings.Contains(bodyTemplate, templateConfig.LeftDelim) {
			renderedBody, err := template.RenderTemplateValues(ctx, bodyTemplate, fmt.Sprintf("%s_%s/bodyTemplate", data.ID, v.reactorName), data.AsMap(), []string{}, templateConfig)
			if err != nil {
				return nil, err
			}
			config.WebhookConfig.Body = string(renderedBody)
		}
	}

	// ===================================================================================
	// Get bearerToken
	// ===================================================================================
	bearerToken, err := v.reactorConfig.Properties["bearerToken"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	config.WebhookConfig.BearerToken = bearerToken

	// ===================================================================================
	// Get signatureHeader
	// ===================================================================================
	signatureHeader, err := v.reactorConfig.Properties["signatureHeader"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	config.WebhookConfig.SignatureHeader = signatureHeader
	if config.WebhookConfig.SignatureHeader == "" {
		config.WebhookConfig.SignatureHeader = settings.SignatureHeader
	}

	// ===================================================================================
	// Get additionalHeaders
	// ===================================================================================
	additionalHeaders, err := v.reactorConfig.Properties["additionalHeaders"].GetMapStringStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	for k, propv := range additionalHeaders {
		if !strings.Contains(propv, templateConfig.LeftDelim) {
			continue
		}
		renderedVal, err := template.RenderTemplateValues(ctx, propv, fmt.Sprintf("%s_%s/additionalHeaders/%s", data.ID, v.reactorName, k), data.AsMap(), []string{}, templateConfig)
		if err != nil {
			return nil, err
		}
		additionalHeaders[k] = string(renderedVal)
	}
	config.WebhookConfig.AdditionalHeaders = additionalHeaders

	// ===================================================================================
	// Get maxRetries
	// ===================================================================================
	maxRetries := 4
	maxRetriesStr, err := v.reactorConfig.Properties["maxRetries"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if maxRetriesStr != "" {
		maxRetries, err = strconv.Atoi(maxRetriesStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied parameters '%v' to an integer. Error: %v", maxRetriesStr, err)
		}
	}
	config.WebhookConfig.MaxRetries = maxRetries

	return config, nil
}

func (v *Reactor) GetHelp() string {
	return reactor.GetReactorHelp(v)
}

func (v *Reactor) GetProperties() []config.ReactorConfigProperty {
	return []config.ReactorConfigProperty{
		{
			Name:        "url",
			Description: "The url to send the webhook to. This field supports go templating",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "webhookSecret",
			Description: "Set the webhook secret to ensure that your server only processes webhook deliveries that were sent by Event Reactor and to ensure that the delivery was not tampered with, you should validate the webhook signature before processing the delivery further. This will help you avoid spending server time to process deliveries that are not from Event Reactor and will help avoid man-in-the-middle attacks.",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "bodyTemplate",
			Description: "The body to send to the webhook. if blank, the original event data will be sent. This field supports go templating.",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "bearerToken",
			Description: "The bearer token to use for authentication. If set, the token will be sent in the Authorization header.",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "signatureHeader",
			Description: "The header to use for the webhook signature. The webhook signature will be singed using sha256. If set, the signature will be sent in this header. Default is X-Event-Reactor-Signature",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "additionalHeaders",
			Description: "Additional headers to send with the webhook. The headers should be in the format of a key value pair. The header value supports go templating.",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeMapString,
		},
		{
			Name:        "maxRetries",
			Description: "The maximum number of times to retry the webhook. Default is 3",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
	}

}

func (v *Reactor) GetRequiredPropertyNames() []string {
	return reactor.GetRequiredPropertyNames(v)
}
