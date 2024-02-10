package powershell

import (
	"context"
	"reflect"
	"testing"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/message"
	"github.com/kcloutie/event-reactor/pkg/pwsh"
	"go.uber.org/zap"
)

func TestReactor_GetProperties(t *testing.T) {
	tests := []struct {
		name string
		v    *Reactor
		want []config.ReactorConfigProperty
	}{
		{
			name: "Test GetProperties",
			v:    &Reactor{},
			want: []config.ReactorConfigProperty{
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v.GetProperties(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reactor.GetProperties() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReactor_GetReactorConfig(t *testing.T) {
	tests := []struct {
		name    string
		v       *Reactor
		wantErr bool
	}{
		{
			name: "Test with valid parameters",
			v: &Reactor{
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"jsonDepth": {
							Value: "4",
						},
						"command": {
							Value: "test-command",
						},
						"parameters": {
							Value: map[string]interface{}{
								"param1": "value1",
								"param2": "value2",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Test with invalid jsonDepth",
			v: &Reactor{
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"jsonDepth": {
							Value: "invalid",
						},
						"command": {
							Value: "test-command",
						},
						"parameters": {
							Value: map[string]interface{}{
								"param1": "value1",
								"param2": "value2",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Test with empty command",
			v: &Reactor{
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"jsonDepth": {
							Value: "4",
						},
						"command": {
							Value: "",
						},
						"parameters": {
							Value: map[string]interface{}{
								"param1": "value1",
								"param2": "value2",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.v.GetReactorConfig(context.Background(), &message.EventData{}, zap.NewNop())
			if (err != nil) != tt.wantErr {
				t.Errorf("Reactor.GetReactorConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReactor_ProcessEvent(t *testing.T) {

	cfg := pwsh.NewPwshExecConfig()
	_, err := cfg.ExecuteRaw(context.Background(), "Get-Help")
	if err != nil {
		// Skip the test if pwsh is not installed
		t.Skipf("Skipping test because pwsh is not installed")
	}
	tests := []struct {
		name    string
		v       *Reactor
		data    *message.EventData
		wantErr bool
	}{
		{
			name: "Test with valid parameters",
			v: &Reactor{
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"jsonDepth": {
							Value: "4",
						},
						"command": {
							Value: "get-help",
						},
						"parameters": {
							Value: map[string]interface{}{
								"Name": "Get-Process",
							},
						},
					},
				},
				Log: zap.NewNop(),
			},
			data:    &message.EventData{},
			wantErr: false,
		},
		{
			name: "Test with missing required properties",
			v: &Reactor{
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"jsonDepth": {
							Value: "4",
						},
						"parameters": {
							Value: map[string]interface{}{
								"param1": "value1",
								"param2": "value2",
							},
						},
					},
				},
				Log: zap.NewNop(),
			},
			data:    &message.EventData{},
			wantErr: true,
		},
		{
			name: "Test with invalid jsonDepth",
			v: &Reactor{
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"jsonDepth": {
							Value: "invalid",
						},
						"command": {
							Value: "test-command",
						},
						"parameters": {
							Value: map[string]interface{}{
								"param1": "value1",
								"param2": "value2",
							},
						},
					},
				},
				Log: zap.NewNop(),
			},
			data:    &message.EventData{},
			wantErr: true,
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.v.ProcessEvent(context.Background(), tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reactor.ProcessEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
