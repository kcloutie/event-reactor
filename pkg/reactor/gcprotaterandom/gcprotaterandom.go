package gcprotaterandom

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/gcp"
	"github.com/kcloutie/event-reactor/pkg/message"
	"github.com/kcloutie/event-reactor/pkg/password"
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
	PasswordLength         int
	UseLowerCase           bool
	UseUpperCase           bool
	UseSpecial             bool
	UseNumbers             bool
	UseExistingSecretValue bool
	SpecialCharOverride    string
	GcpProject             string
	GcpSecretName          string
}

func New() *Reactor {
	return &Reactor{
		reactorName: "gcp/secret/rotate/random",
	}
}

func (v *Reactor) SetLogger(logger *zap.Logger) {
	v.Log = logger
}
func (v *Reactor) GetName() string {
	return v.reactorName
}

func (v *Reactor) GetDescription() string {
	return "This reactor is used to rotate a GCP secret with a random value. The random value is generated using the crypto/rand package. The random value is then used to update the secret in GCP Secret Manager."
}

func (v *Reactor) SetReactor(reactor config.ReactorConfig) {
	v.reactorConfig = reactor
}

func (v *Reactor) GetConfigExample() string {
	return `
- name: test_gcpRotateRandom
celExpressionFilter: has(attributes.eventType) && attributes['eventType'] == 'SECRET_ROTATE'
failOnError: false
disabled: false
type: gcp/secret/rotate/random
properties:
  useLowerCase:
    value: true
  useUpperCase:
    value: true
  useSpecial:
    value: true
  useNumbers:
    value: true
  secretFullName:
    payloadValue:
      propertyPaths:
      - data.name`
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
	password := password.GeneratePassword(reactorConfig.PasswordLength, reactorConfig.UseLowerCase, reactorConfig.UseUpperCase, reactorConfig.UseSpecial, reactorConfig.UseNumbers, reactorConfig.SpecialCharOverride)

	if reactorConfig.UseExistingSecretValue {
		v.Log.Info("Using the existing value of the secret to update the secret with")
		password, err = gcp.GetSecret(ctx, nil, reactorConfig.GcpProject, reactorConfig.GcpSecretName, "latest")
		if err != nil {
			return fmt.Errorf("failed to get the value of the GCP secret: %v", err)
		}
	}

	version, err := gcp.AddVersion(ctx, nil, reactorConfig.GcpProject, reactorConfig.GcpSecretName, password)
	if err != nil {
		return fmt.Errorf("failed to add a new version to the GCP secret: %v", err)
	}
	v.Log.Info("Successfully added a new version to the GCP secret", zap.String("version", version))

	return nil
}

func (v *Reactor) GetReactorConfig(ctx context.Context, data *message.EventData, log *zap.Logger) (*ReactorConfig, error) {

	config := &ReactorConfig{}

	passwordLength := 20
	passwordLengthStr, err := v.reactorConfig.Properties["passwordLength"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if passwordLengthStr != "" {
		passwordLength, err = strconv.Atoi(passwordLengthStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied parameters '%v' to an integer. Error: %v", passwordLengthStr, err)
		}
	}
	config.PasswordLength = passwordLength

	specialCharOverride, err := v.reactorConfig.Properties["specialCharOverride"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	config.SpecialCharOverride = specialCharOverride

	config.UseLowerCase = true
	useLowerCaseStr, err := v.reactorConfig.Properties["useLowerCase"].GetStringValue(ctx, v.Log, data)
	if err == nil && useLowerCaseStr != "" {
		config.UseLowerCase, err = strconv.ParseBool(useLowerCaseStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied useLowerCase '%v' to a boolean. Error: %v", useLowerCaseStr, err)
		}
	}

	config.UseUpperCase = true
	useUpperCaseStr, err := v.reactorConfig.Properties["useUpperCase"].GetStringValue(ctx, v.Log, data)
	if err == nil && useUpperCaseStr != "" {
		config.UseUpperCase, err = strconv.ParseBool(useUpperCaseStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied useUpperCase '%v' to a boolean. Error: %v", useUpperCaseStr, err)
		}
	}

	config.UseSpecial = true
	useSpecialStr, err := v.reactorConfig.Properties["useSpecial"].GetStringValue(ctx, v.Log, data)
	if err == nil && useSpecialStr != "" {
		config.UseSpecial, err = strconv.ParseBool(useSpecialStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied useSpecial '%v' to a boolean. Error: %v", useSpecialStr, err)
		}
	}

	config.UseExistingSecretValue = true
	useExistingSecretValueStr, err := v.reactorConfig.Properties["useExistingSecretValue"].GetStringValue(ctx, v.Log, data)
	if err == nil && useExistingSecretValueStr != "" {
		config.UseExistingSecretValue, err = strconv.ParseBool(useExistingSecretValueStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied useExistingSecretValue '%v' to a boolean. Error: %v", useExistingSecretValueStr, err)
		}
	}

	config.UseNumbers = true
	useNumbersStr, err := v.reactorConfig.Properties["useNumbers"].GetStringValue(ctx, v.Log, data)
	if err == nil && useNumbersStr != "" {
		config.UseNumbers, err = strconv.ParseBool(useNumbersStr)
		if err != nil {
			return nil, fmt.Errorf("failed to convert the supplied useNumbers '%v' to a boolean. Error: %v", useNumbersStr, err)
		}
	}

	// secretFullName

	secretFullName, err := v.reactorConfig.Properties["secretFullName"].GetStringValue(ctx, v.Log, data)
	if err != nil {
		return nil, err
	}
	if secretFullName == "" {

		project, err := v.reactorConfig.Properties["project"].GetStringValue(ctx, v.Log, data)
		if err != nil {
			return nil, err
		}
		if project == "" {
			return nil, fmt.Errorf("the project property was not supplied or was empty")
		}
		config.GcpProject = project

		secretName, err := v.reactorConfig.Properties["secretName"].GetStringValue(ctx, v.Log, data)
		if err != nil {
			return nil, err
		}
		if secretName == "" {
			return nil, fmt.Errorf("the secretName property was not supplied or was empty")
		}
		config.GcpSecretName = secretName

	} else {
		secretFullNameArr := strings.Split(secretFullName, "/")
		if len(secretFullNameArr) != 4 {
			return nil, fmt.Errorf("the secretFullName property was not in the correct format. The correct format is projects/<PROJECT_ID>/secrets/<SECRET_NAME>")
		}
		config.GcpProject = secretFullNameArr[1]
		config.GcpSecretName = secretFullNameArr[3]

		if config.GcpSecretName == "" {
			return nil, fmt.Errorf("the secretName property was not supplied or was empty")
		}
		if config.GcpProject == "" {
			return nil, fmt.Errorf("the project property was not supplied or was empty")
		}
	}
	return config, nil
}

func (v *Reactor) GetProperties() []config.ReactorConfigProperty {
	return []config.ReactorConfigProperty{
		{
			Name:        "passwordLength",
			Description: "The length of the random password to generate",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "useLowerCase",
			Description: "Use lower case letters in the random password",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "useUpperCase",
			Description: "Use upper case letters in the random password",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "useSpecial",
			Description: "Use special characters in the random password",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "useNumbers",
			Description: "Use numbers in the random password",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "specialCharOverride",
			Description: "Override the default special characters to use in the random password",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "project",
			Description: "The GCP project where the secret is located",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "secretName",
			Description: "The name of the GCP secret to rotate",
			Required:    config.AsBoolPointer(false),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "secretFullName",
			Description: "The full name of the GCP secret to rotate i.e projects/<PROJRCT_ID>/secrets/<SECRET_NAME>. This property is used to override the project and secretName properties. If this property is supplied, the project and secretName properties will be ignored",
			Required:    config.AsBoolPointer(true),
			Type:        config.PropertyTypeString,
		},
		{
			Name:        "useExistingSecretValue",
			Description: "Use the existing value of the secret to update the secret with. If this property is set to true, the passwordLength, useLowerCase, useUpperCase, useSpecial, useNumbers, and specialCharOverride properties will be ignored",
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
