package gcppublishpubsub

import (
	"context"
	"fmt"
	"strings"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/gcp"
	"github.com/kcloutie/event-reactor/pkg/message"
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
	Project    string
	TopicId    string
	Payload    string
	Attributes map[string]string
}

func New() *Reactor {
	return &Reactor{
		reactorName: "gcp/pubsub/publish/message",
	}
}

func (v *Reactor) SetLogger(logger *zap.Logger) {
	v.Log = logger
}
func (v *Reactor) GetName() string {
	return v.reactorName
}

func (v *Reactor) GetDescription() string {
	return "This reactor is used to publish a message to a GCP Pub/Sub topic. The message is published to the topic using the GCP SDK. The payload can be a Go template. The Go template can use the data, attributes, and id properties of the event data. In addition the attributes key value pairs can be Go templates."
}

func (v *Reactor) SetReactor(reactor config.ReactorConfig) {
	v.reactorConfig = reactor
}

func (v *Reactor) GetConfigExample() string {
	return `
- name: test_gcpPublishPubSub
  celExpressionFilter: has(attributes.eventType) && attributes['eventType'] == 'SECRET_VERSION_ADD'
  failOnError: false
  disabled: false
  type: gcp/pubsub/publish/message
  properties:
    topicId:
      value: some-topic
    project:
      value: some-project
    payload:
      value: '{"appName":"testApp","platform":"cloudRun"}'
    attributes:
      value:
        eventType: "REDEPLOY_APP"
        dataFormat: "JSON_API_V1"
        secretId: "{{ .attributes.secretId }}"
`
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

	messageId, err := gcp.PublishEvent(reactorConfig.Project, reactorConfig.TopicId, []byte(reactorConfig.Payload), reactorConfig.Attributes)
	if err != nil {
		return err
	}
	v.Log.Info("Successfully published message to pub/sub", zap.String("messageId", messageId), zap.String("topicId", reactorConfig.TopicId), zap.String("project", reactorConfig.Project))

	return nil
}

func (v *Reactor) GetReactorConfig(ctx context.Context, data *message.EventData, log *zap.Logger) (*ReactorConfig, error) {

	config := &ReactorConfig{}

	templateConfig := template.NewRenderTemplateOptions()
	reactor.SetGoTemplateOptionValues(ctx, v.Log, &templateConfig, v.reactorConfig.Properties)

	// ===================================================================================
	// Get Project
	// ===================================================================================
	project, err := v.reactorConfig.Properties["project"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if project == "" {
		return nil, fmt.Errorf("the project property was not supplied or was empty")
	}
	config.Project = project

	// ===================================================================================
	// Get Topic Id
	// ===================================================================================
	topicId, err := v.reactorConfig.Properties["topicId"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if topicId == "" {
		return nil, fmt.Errorf("the topicId property was not supplied or was empty")
	}
	config.TopicId = topicId

	// ===================================================================================
	// Get Payload
	// ===================================================================================
	payload, err := v.reactorConfig.Properties["payload"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	// if payload == "" {
	// 	return nil, fmt.Errorf("the payload property was not supplied or was empty")
	// }
	renderedPayload := []byte(payload)
	if strings.Contains(payload, "{{") {
		renderedPayload, err = template.RenderTemplateValues(ctx, payload, fmt.Sprintf("%s_%s/payload", data.ID, v.reactorName), data.AsMap(), []string{}, templateConfig)
		if err != nil {
			return nil, err
		}
	}
	config.Payload = string(renderedPayload)

	// ===================================================================================
	// Get Attributes
	// ===================================================================================

	attributes, err := v.reactorConfig.Properties["attributes"].GetMapStringStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}

	for k, val := range attributes {
		renderedAttributeVal, err := template.RenderTemplateValues(ctx, val, fmt.Sprintf("%s_%s/param_%s", data.ID, v.reactorName, k), data.AsMap(), []string{}, templateConfig)
		if err != nil {
			return nil, err
		}
		attributes[k] = string(renderedAttributeVal)

	}
	config.Attributes = attributes

	return config, nil
}

func (v *Reactor) GetProperties() []config.ReactorConfigProperty {
	return []config.ReactorConfigProperty{
		{
			Name:        "project",
			Description: "The GCP project id where the pub/sub topic is located",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "topicId",
			Description: "The id of the pub/sub topic",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "payload",
			Description: "The payload to publish to the pub/sub topic. This should be json and supports Go templates",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "attributes",
			Description: "The attributes to publish to the pub/sub topic. This should be a map of key value pairs and supports Go templates",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeMapString,
		},
	}
}

func (v *Reactor) GetRequiredPropertyNames() []string {
	return reactor.GetRequiredPropertyNames(v)
}

func (v *Reactor) GetHelp() string {
	return reactor.GetReactorHelp(v)
}
