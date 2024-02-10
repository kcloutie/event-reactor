package email

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kcloutie/event-reactor/pkg/config"
	em "github.com/kcloutie/event-reactor/pkg/email"
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
	EmailCOnfig em.EmailConfiguration
}

func New() *Reactor {
	return &Reactor{
		reactorName: "email",
	}
}

func (v *Reactor) GetConfigExample() string {
	return `
  reactorConfigs:
  - name: test_email
    celExpressionFilter: attributes.test == 'email'
    type: email
    properties:
      subject:
        value: Testing email
      from:
        value: someone@somewhere.com
      password:
        valueFrom:
          secretKeyRef:
            name: some_secret_for_email
            projectId: some-gcp-project
            version: latest
      to:
        value: someoneelse@somewherelese.com
      body:
        value: this is a test email from event reactor
      smtpHost:
        value: smpt.com
      smtpPort:
        value: "587"
      maxRetries:
        value: "5"
`
}

func (v *Reactor) SetLogger(logger *zap.Logger) {
	v.Log = logger
}
func (v *Reactor) GetName() string {
	return v.reactorName
}

func (v *Reactor) GetDescription() string {
	return "This reactor sends an email to the specified recipient(s) using the supplied smtp server and credentials. The email subject and body can be templated using Go's text/template package. The email is retried up to the specified number of times if it fails to send."
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
	v.Log = v.Log.With(zap.String("from", reactorConfig.EmailCOnfig.From), zap.String("smtpHost", reactorConfig.EmailCOnfig.SMTPHost), zap.Int("smtpPort", reactorConfig.EmailCOnfig.SMTPPort), zap.Strings("to", reactorConfig.EmailCOnfig.To), zap.String("subject", reactorConfig.EmailCOnfig.Subject))
	v.Log.Debug("Sending email")
	err = reactorConfig.EmailCOnfig.SendEmail(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (v *Reactor) GetReactorConfig(ctx context.Context, data *message.EventData, log *zap.Logger) (*ReactorConfig, error) {

	config := ReactorConfig{
		EmailCOnfig: em.EmailConfiguration{
			SleepInterval: time.Duration(1) * time.Second,
			Log:           v.Log,
		},
	}

	templateConfig := template.NewRenderTemplateOptions()
	reactor.SetGoTemplateOptionValues(ctx, v.Log, &templateConfig, v.reactorConfig.Properties)

	// ===================================================================================
	// Get smtpHost
	// ===================================================================================
	smtpHost, err := v.reactorConfig.Properties["smtpHost"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if smtpHost == "" {
		return nil, fmt.Errorf("the smtpHost property was not supplied or was empty")
	}
	config.EmailCOnfig.SMTPHost = smtpHost

	// ===================================================================================
	// Get smtpPort
	// ===================================================================================

	smtpPortStr, err := v.reactorConfig.Properties["smtpPort"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if smtpPortStr == "" {
		return nil, fmt.Errorf("the smtpPort property was not supplied or was empty")
	}
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {

		return nil, fmt.Errorf("failed to convert the supplied smtpPort '%v' to an integer. Error: %v", smtpPortStr, err)
	}
	config.EmailCOnfig.SMTPPort = smtpPort

	// ===================================================================================
	// Get maxRetries
	// ===================================================================================

	maxRetries := 5
	maxRetriesStr, err := v.reactorConfig.Properties["maxRetries"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if maxRetriesStr != "" {
		maxRetries, err = strconv.Atoi(maxRetriesStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied maxRetries '%v' to an integer. Error: %v", maxRetriesStr, err)
		}
	}

	config.EmailCOnfig.MaxRetries = maxRetries

	// ===================================================================================
	// Get from
	// ===================================================================================

	from, err := v.reactorConfig.Properties["from"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if from == "" {
		return nil, fmt.Errorf("the from property was not supplied or was empty")
	}
	config.EmailCOnfig.From = from

	// ===================================================================================
	// Get to
	// ===================================================================================

	to, err := v.reactorConfig.Properties["to"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if to == "" {
		return nil, fmt.Errorf("the to property was not supplied or was empty")
	}
	config.EmailCOnfig.To = splitEmailAddress(to)

	// ===================================================================================
	// Get password
	// ===================================================================================

	password, err := v.reactorConfig.Properties["password"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if from == "" {
		return nil, fmt.Errorf("the password property was not supplied or was empty")
	}
	config.EmailCOnfig.Password = password

	// ===================================================================================
	// Get subject
	// ===================================================================================

	subject, err := v.reactorConfig.Properties["subject"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	renderedSubject, err := template.RenderTemplateValues(ctx, subject, fmt.Sprintf("%s_%s/subject", data.ID, v.reactorName), data.AsMap(), []string{}, templateConfig)
	if err != nil {
		return nil, err
	}
	config.EmailCOnfig.Subject = string(renderedSubject)

	// ===================================================================================
	// Get body
	// ===================================================================================

	body, err := v.reactorConfig.Properties["body"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	renderedBody, err := template.RenderTemplateValues(ctx, body, fmt.Sprintf("%s_%s/body", data.ID, v.reactorName), data.AsMap(), []string{}, templateConfig)
	if err != nil {
		return nil, err
	}
	if body == "" {
		return nil, fmt.Errorf("the body property was not supplied or was empty")
	}
	config.EmailCOnfig.Subject = string(renderedBody)

	return &config, nil

}

func (v *Reactor) GetHelp() string {
	return reactor.GetReactorHelp(v)
}

func (v *Reactor) GetProperties() []config.ReactorConfigProperty {
	return []config.ReactorConfigProperty{
		{
			Name:        "from",
			Description: "The email address of the sender",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "password",
			Description: "The password for the smtp server",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "to",
			Description: "The email address of the recipient(s). Multiple addresses can be separated by a comma, semicolon, or space",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "subject",
			Description: "The subject of the email. This field supports go templating",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "body",
			Description: "The body of the email. This field supports go templating",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "smtpHost",
			Description: "The smtp server host",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "smtpPort",
			Description: "The smtp server port",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "maxRetries",
			Description: "The maximum number of times to retry sending the email. Defaults to 5",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
	}
}

func (v *Reactor) GetRequiredPropertyNames() []string {
	return reactor.GetRequiredPropertyNames(v)
}

func splitEmailAddress(address string) []string {

	if strings.Contains(strings.Trim(address, " "), ";") {
		return strings.Split(address, ";")
	}

	if strings.Contains(strings.Trim(address, " "), ",") {
		return strings.Split(address, ",")
	}

	if strings.Contains(strings.Trim(address, " "), " ") {
		return strings.Split(address, " ")
	}
	return []string{address}
}
