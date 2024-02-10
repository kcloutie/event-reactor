package reactor

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/template"
	"go.uber.org/zap/zaptest"
)

func TestHasRequiredProperties(t *testing.T) {
	val1 := "value1"
	val2 := "value2"
	val3 := "value3"
	type args struct {
		properties         map[string]config.PropertyAndValue
		requiredProperties []string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "All required properties are present",
			args: args{
				properties: map[string]config.PropertyAndValue{
					"prop1": {Value: &val1},
					"prop2": {Value: &val2},
					"prop3": {Value: &val3},
				},
				requiredProperties: []string{"prop1", "prop2", "prop3"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Some required properties are missing",
			args: args{
				properties: map[string]config.PropertyAndValue{
					"prop1": {Value: &val1},
					"prop2": {Value: &val2},
				},
				requiredProperties: []string{"prop1", "prop2", "prop3"},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "All required properties are missing",
			args: args{
				properties: map[string]config.PropertyAndValue{
					"prop1": {Value: &val1},
					"prop2": {Value: &val2},
				},
				requiredProperties: []string{"prop3", "prop4"},
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HasRequiredProperties(tt.args.properties, tt.args.requiredProperties)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasRequiredProperties() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HasRequiredProperties() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetGoTemplateOptionValues(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	type args struct {
		properties map[string]config.PropertyAndValue
	}
	tests := []struct {
		name string
		args args
		want template.RenderTemplateOptions
	}{
		{
			name: "Properties exist and GetValue does not return an error",
			args: args{
				properties: map[string]config.PropertyAndValue{
					template.LeftDelimPropertyName:  {Value: strPtr("||")},
					template.RightDelimPropertyName: {Value: strPtr("}}")},
					template.IgnoreTemplateErrors:   {Value: strPtr("true")},
				},
			},
			want: template.RenderTemplateOptions{
				LeftDelim:            "||",
				RightDelim:           "}}",
				IgnoreTemplateErrors: false,
			},
		},
		{
			name: "Properties do not exist",
			args: args{
				properties: map[string]config.PropertyAndValue{},
			},
			want: template.RenderTemplateOptions{
				LeftDelim:            "{{",
				RightDelim:           "}}",
				IgnoreTemplateErrors: false,
			},
		},
		{
			name: "invalid value for IgnoreTemplateErrors property",
			args: args{
				properties: map[string]config.PropertyAndValue{
					template.IgnoreTemplateErrors: {Value: strPtr("notbool")},
				},
			},
			want: template.RenderTemplateOptions{
				LeftDelim:            "{{",
				RightDelim:           "}}",
				IgnoreTemplateErrors: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &template.RenderTemplateOptions{}
			SetGoTemplateOptionValues(ctx, logger, config, tt.args.properties)

			// Add assertions here to verify the expected state of `config` after calling the function
		})
	}
}

// strPtr is a helper function for creating a pointer to a string
func strPtr(s string) *string {
	return &s
}

// boolPtr is a helper function for creating a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}

func TestGetRequiredPropertyNames(t *testing.T) {
	type fields struct {
		Properties []config.ReactorConfigProperty
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "No properties are required",
			fields: fields{
				Properties: []config.ReactorConfigProperty{
					{Name: "prop1", Required: boolPtr(false)},
					{Name: "prop2", Required: boolPtr(false)},
					{Name: "prop3", Required: boolPtr(false)},
				},
			},
			want: []string{},
		},
		{
			name: "Some properties are required",
			fields: fields{
				Properties: []config.ReactorConfigProperty{
					{Name: "prop1", Required: boolPtr(true)},
					{Name: "prop2", Required: boolPtr(false)},
					{Name: "prop3", Required: boolPtr(true)},
				},
			},
			want: []string{"prop1", "prop3"},
		},
		{
			name: "All properties are required",
			fields: fields{
				Properties: []config.ReactorConfigProperty{
					{Name: "prop1", Required: boolPtr(true)},
					{Name: "prop2", Required: boolPtr(true)},
					{Name: "prop3", Required: boolPtr(true)},
				},
			},
			want: []string{"prop1", "prop2", "prop3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewTestReactor()
			r.Properties = tt.fields.Properties
			if got := GetRequiredPropertyNames(r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRequiredPropertyNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

// intPtr is a helper function for creating a pointer to an int
func intPtr(i int) *int {
	return &i
}

func TestGetReactorHelp(t *testing.T) {
	type args struct {
		p ReactorInterface
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Test with a reactor with properties",
			args: args{
				p: &Reactor{
					Name:        "Test Reactor",
					Description: "This is a test reactor",
					Properties: []config.ReactorConfigProperty{
						{
							Name:        "prop1",
							Description: "This is property 1",
							Required:    boolPtr(true),
							Type:        "string",
							Validation: &config.PropertyValidation{
								ValidationRegex:        "^test$",
								ValidationRegexMessage: "Must match ^test$",
								AllowNullOrEmpty:       boolPtr(false),
								MinLength:              intPtr(1),
								MaxLength:              intPtr(10),
							},
						},
					},
				},
			},
			want: `NAME
  Test Reactor

DESCRIPTION
  This is a test reactor

PROPERTIES

  prop1
    This is property 1

    Required:                 true
    Type:                     string
    Validation Regex:         ^test$
    Validation Regex Message: Must match ^test$
    Allow Null/Empty:         false
    Min Length:               1
    Max Length:               10


EXAMPLE CONFIG

  

`,
		},
		{
			name: "Test with a reactor without properties",
			args: args{
				p: &Reactor{
					Name:        "Test Reactor",
					Description: "This is a test reactor",
					Properties:  []config.ReactorConfigProperty{},
				},
			},
			want: `NAME
  Test Reactor

DESCRIPTION
  This is a test reactor

PROPERTIES

EXAMPLE CONFIG

  

`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetReactorHelp(tt.args.p)

			os.WriteFile("got.txt", []byte(got), 0644)
			os.WriteFile("want.txt", []byte(tt.want), 0644)

			if !strings.Contains(tt.want, got) {
				t.Errorf("GetReactorHelp() = %v, want %v", got, tt.want)
			}
		})
	}
}
