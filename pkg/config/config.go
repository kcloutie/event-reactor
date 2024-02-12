package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kcloutie/event-reactor/pkg/gcp"
	"github.com/kcloutie/event-reactor/pkg/maps"
	"github.com/kcloutie/event-reactor/pkg/message"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type PropertyTypes string

const (
	PropertyTypeString      PropertyTypes = "string"
	PropertyTypeStringArray PropertyTypes = "stringArray"
	PropertyTypeMapString   PropertyTypes = "mapString"
	// PropertyTypeMapStringString  PropertyTypes = "mapStringString"
	// PropertyTypeInt  PropertyTypes = "int"
	// PropertyTypeBool  PropertyTypes = "bool"
	// PropertyTypeFloat  PropertyTypes = "float"
)

func AsBoolPointer(val bool) *bool {
	return &val
}

func AsStringPointer(val string) *string {
	return &val
}

type ServerConfiguration struct {
	ReactorConfigs      []ReactorConfig `json:"reactorConfigs,omitempty" yaml:"reactorConfigs,omitempty"`
	TraceHeaderKey      string          `json:"traceHeaderKey,omitempty" yaml:"traceHeaderKey,omitempty"`
	LoadTestReactor     bool            `json:"loadTestReactor,omitempty" yaml:"loadTestReactor,omitempty"`
	AlwaysReturn200     bool            `json:"alwaysReturn200,omitempty" yaml:"alwaysReturn200,omitempty"`
	LogRawPubSubPayload bool            `json:"logRawPubSubPayload,omitempty" yaml:"logRawPubSubPayload,omitempty"`
	LogEventDataPayload bool            `json:"logEventDataPayload,omitempty" yaml:"logEventDataPayload,omitempty"`
	//X-Cloud-Trace-Context
}

type ReactorConfig struct {
	Name                string                      `json:"name,omitempty" yaml:"name,omitempty"`
	CelExpressionFilter string                      `json:"celExpressionFilter,omitempty" yaml:"celExpressionFilter,omitempty"`
	Disabled            bool                        `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	Type                string                      `json:"type,omitempty" yaml:"type,omitempty"`
	Properties          map[string]PropertyAndValue `json:"properties,omitempty" yaml:"properties,omitempty"`
	FailOnError         *bool                       `json:"failOnError,omitempty" yaml:"failOnError,omitempty"`
}

func (rc *ReactorConfig) GetFailOnError() bool {
	if rc.FailOnError == nil {
		return true
	}
	return *rc.FailOnError
}

type PropertyAndValue struct {
	// Name         string               `json:"name,omitempty" yaml:"name,omitempty"`
	Value        interface{}          `json:"value,omitempty" yaml:"value,omitempty"`
	ValueFrom    *PropertyValueSource `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
	PayloadValue *PayloadValueRef     `json:"payloadValue,omitempty" yaml:"payloadValue,omitempty"`
	FromFile     *string              `json:"fromFile,omitempty" yaml:"fromFile,omitempty"`
	FromEnv      *string              `json:"fromEnv,omitempty" yaml:"fromEnv,omitempty"`
}

type ReactorConfigProperty struct {
	Name        string              `json:"name" yaml:"name"`
	Required    *bool               `json:"required" yaml:"required"`
	Description string              `json:"description" yaml:"description"`
	Type        PropertyTypes       `json:"type" yaml:"type"`
	Validation  *PropertyValidation `json:"validation" yaml:"validation"`
}

type PropertyValidation struct {
	ValidationRegex        string `json:"validationRegex,omitempty" yaml:"validationRegex,omitempty"`
	ValidationRegexMessage string `json:"validationRegexMessage,omitempty" yaml:"validationRegexMessage,omitempty"`
	AllowNullOrEmpty       *bool  `json:"allowNullOrEmpty,omitempty" yaml:"allowNullOrEmpty,omitempty"`
	MinLength              *int   `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength              *int   `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
}

// var templateConfig = template.NewRenderTemplateOptions()

func (pv *PropertyAndValue) GetValueProp(ctx context.Context, data *message.EventData) (interface{}, error) {
	if pv.Value != nil {
		switch v := pv.Value.(type) {
		case *string:
			// if strings.Contains(*v, "{{") {
			// 	rendered, err := template.RenderTemplateValues(ctx, *v, "GetValueProp/*string", data.AsMap(), []string{}, templateConfig)
			// 	if err != nil {
			// 		return nil, fmt.Errorf("error rendering template for property value %s: %w", *v, err)
			// 	}
			// 	return string(rendered), nil
			// }
			return *v, nil
		case *[]string:
			return *v, nil
		case *map[string]interface{}:
			return *v, nil
		case *map[string]string:
			return *v, nil
		case *interface{}:
			return *v, nil
		// case string:
		// 	if strings.Contains(v, "{{") {
		// 		rendered, err := template.RenderTemplateValues(ctx, v, "GetValueProp/string", data.AsMap(), []string{}, templateConfig)
		// 		if err != nil {
		// 			return nil, fmt.Errorf("error rendering template for property value %s: %w", v, err)
		// 		}
		// 		return string(rendered), nil
		// 	}
		default:
			return v, nil
		}
	}
	return nil, nil
}

func (pv *PropertyAndValue) GetFromFileProp() string {
	if pv.FromFile != nil {
		return *pv.FromFile
	}
	return ""
}

func (pv *PropertyAndValue) GetFromEnvProp() string {
	if pv.FromEnv != nil {
		return *pv.FromEnv
	}
	return ""
}

type PropertyValueSource struct {
	GcpSecretRef *GcpSecretRef `json:"secretKeyRef,omitempty" yaml:"secretKeyRef,omitempty"`
}

type GcpSecretRef struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	ProjectId string `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	Version   string `json:"version,omitempty" yaml:"version,omitempty"`
	Path      string `json:"path,omitempty" yaml:"path,omitempty"`
}

type PayloadValueRef struct {
	PropertyPaths []string `json:"propertyPaths,omitempty" yaml:"propertyPaths,omitempty"`
}

func (o PropertyAndValue) GetStringValue(ctx context.Context, log *zap.Logger, data *message.EventData) (string, error) {
	val, err := o.GetValue(ctx, log, data)
	if err != nil {
		return "", err
	}
	switch v := val.(type) {
	case string:
		return v, nil
	default:
		if v == nil {
			return "", nil
		}
		return "", fmt.Errorf("expected the value to be the type of, however it is of type %T", val)
	}
}

func (o PropertyAndValue) GetStringArrayValue(ctx context.Context, log *zap.Logger, data *message.EventData) ([]string, error) {
	val, err := o.GetValue(ctx, log, data)
	if err != nil {
		return []string{}, err
	}
	switch v := val.(type) {
	case []string:
		return v, nil
	default:
		if v == nil {
			return []string{}, nil
		}
		return nil, fmt.Errorf("expected the value to be the type of, however it is of type %T", val)
	}
}

func (o PropertyAndValue) GetMapStringStringValue(ctx context.Context, log *zap.Logger, data *message.EventData) (map[string]string, error) {
	val, err := o.GetValue(ctx, log, data)
	if err != nil {
		return nil, err
	}
	switch v := val.(type) {
	case map[string]string:
		return v, nil
	case map[string]interface{}:
		results := map[string]string{}
		for k, v := range v {
			results[k] = fmt.Sprintf("%v", v)
		}
		return results, nil
	default:
		if v == nil {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("expected the value to be the type of map[string]string, however it is of type %T", val)
	}
}

func (o PropertyAndValue) GetMapStringInterfaceValue(ctx context.Context, log *zap.Logger, data *message.EventData) (map[string]interface{}, error) {
	val, err := o.GetValue(ctx, log, data)
	if err != nil {
		return nil, err
	}
	switch v := val.(type) {
	case map[string]interface{}:
		return v, nil
	case map[string]string:
		newRes := map[string]interface{}{}
		for k, v := range v {
			newRes[k] = v
		}
		return newRes, nil
	default:
		if v == nil {
			return map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("expected the value to be the type of map[string]string, however it is of type %T", val)
	}
}

func (o PropertyAndValue) GetValue(ctx context.Context, log *zap.Logger, data *message.EventData) (interface{}, error) {
	if o.ValueFrom != nil && o.ValueFrom.GcpSecretRef != nil {
		secClient := gcp.FromCtx(ctx)
		if secClient != nil {
			defer secClient.Close()
		}
		val, err := gcp.GetSecret(ctx, secClient, o.ValueFrom.GcpSecretRef.ProjectId, o.ValueFrom.GcpSecretRef.Name, o.ValueFrom.GcpSecretRef.Version)
		if err != nil {
			return "", fmt.Errorf("error getting secret %s/%s/%s: %w", o.ValueFrom.GcpSecretRef.ProjectId, o.ValueFrom.GcpSecretRef.Name, o.ValueFrom.GcpSecretRef.Version, err)
		}
		if o.ValueFrom.GcpSecretRef.Path != "" {
			secretData := map[string]interface{}{}
			err = json.Unmarshal([]byte(val), &secretData)
			if err != nil {
				err = yaml.Unmarshal([]byte(val), &secretData)
				if err != nil {
					return "", fmt.Errorf("failed to unmarshal the secret content using yaml and json. When specifying a path to get the secret value, it is expected the content is either json or yaml - %w", err)
				}
			}
			val, err = maps.GetStringValueFromMapByPath(maps.MapPath(o.ValueFrom.GcpSecretRef.Path), secretData, true)
			if err != nil {
				return "", fmt.Errorf("failed to get the secret from the '%s' path - %w", o.ValueFrom.GcpSecretRef.Path, err)
			}
		}
		return val, nil
	}
	if o.PayloadValue != nil && len(o.PayloadValue.PropertyPaths) != 0 {
		errs := []string{}
		for _, path := range o.PayloadValue.PropertyPaths {
			val, err := data.GetPropertyValue(path)
			if err == nil {
				return val, nil
			} else {
				errs = append(errs, err.Error())
				continue
			}
		}
		return "", fmt.Errorf("error getting property value from the following paths '%s'. Errors: %s", strings.Join(o.PayloadValue.PropertyPaths, ", "), strings.Join(errs, ", "))
	}

	if o.GetFromFileProp() != "" {
		fileBytes, err := os.ReadFile(o.GetFromFileProp())
		if err != nil {
			return "", fmt.Errorf("error reading file %s: %w", o.GetFromFileProp(), err)
		}
		return string(fileBytes), nil
	}

	if o.GetFromEnvProp() != "" {
		val := os.Getenv(o.GetFromEnvProp())
		if val == "" {
			log.Warn(fmt.Sprintf("environment variable %s is empty", o.GetFromEnvProp()))
		}
		log.Debug(fmt.Sprintf("environment variable %s value is %s", o.GetFromEnvProp(), val))
		return os.Getenv(o.GetFromEnvProp()), nil
	}

	return o.GetValueProp(ctx, data)
}

func NewServerConfiguration() *ServerConfiguration {
	return &ServerConfiguration{}
}

var config *ServerConfiguration

type ctxConfigKey struct{}

func FromCtx(ctx context.Context) *ServerConfiguration {
	if l, ok := ctx.Value(ctxConfigKey{}).(*ServerConfiguration); ok {
		return l
	} else if l := config; l != nil {
		return l
	}
	return NewServerConfiguration()
}

func WithCtx(ctx context.Context, l *ServerConfiguration) context.Context {
	if lp, ok := ctx.Value(ctxConfigKey{}).(*ServerConfiguration); ok {
		if lp == l {
			return ctx
		}
	}
	return context.WithValue(ctx, ctxConfigKey{}, l)
}
