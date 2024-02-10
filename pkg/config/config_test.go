package config

import (
	"context"
	"fmt"
	"hash/crc32"
	"os"
	"reflect"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/kcloutie/event-reactor/pkg/gcp"
	"github.com/kcloutie/event-reactor/pkg/message"
	"go.uber.org/zap/zaptest"
)

func TestFromCtx(t *testing.T) {
	cfg := NewServerConfiguration()
	cfg.TraceHeaderKey = "HEADER"
	ctx := context.Background()
	ctx = WithCtx(ctx, cfg)
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want *ServerConfiguration
	}{
		{
			name: "already exists",
			args: args{
				ctx: ctx,
			},
			want: cfg,
		},
		{
			name: "new",
			args: args{
				ctx: context.Background(),
			},
			want: NewServerConfiguration(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromCtx(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromCtx() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetValue(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	testVal := "test value"
	ignoreVal := "ignored"
	fromFilePath := "testdata/fromfile.txt"
	testEnvName := "TEST_ENV"
	data := message.EventData{
		Data: map[string]interface{}{
			"test": "data test value",
			"child1": map[string]interface{}{
				"child2": map[string]interface{}{
					"test_child": "data test child value",
				},
			},
		},
		Attributes: map[string]string{},
		ID:         "test-id",
	}
	tests := []struct {
		name          string
		propVal       PropertyAndValue
		data          *message.EventData
		want          interface{}
		secretContent string
		envValue      string
		wantErr       bool
	}{
		{
			name: "ValueFrom is nil",
			propVal: PropertyAndValue{
				Value: &testVal,
			},
			data:    &data,
			want:    "test value",
			wantErr: false,
		},
		{
			name: "string value",
			propVal: PropertyAndValue{
				Value: "string value",
			},
			data:    &data,
			want:    "string value",
			wantErr: false,
		},
		{
			name: "string array value",
			propVal: PropertyAndValue{
				Value: []string{"string value"},
			},
			data:    &data,
			want:    []string{"string value"},
			wantErr: false,
		},
		{
			name: "map string ptr value",
			propVal: PropertyAndValue{
				Value: &map[string]interface{}{
					"test": "string value",
				},
			},
			data: &data,
			want: map[string]interface{}{
				"test": "string value",
			},
			wantErr: false,
		},
		{
			name: "ValueFrom is not nil but GcpSecretRef is nil",
			propVal: PropertyAndValue{
				Value: &testVal,
				ValueFrom: &PropertyValueSource{
					GcpSecretRef: nil,
				},
			},
			data:          &data,
			want:          "test value",
			secretContent: "test value",
			wantErr:       false,
		},
		{
			name: "ValueFrom and GcpSecretRef are not nil",
			propVal: PropertyAndValue{
				Value: &testVal,
				ValueFrom: &PropertyValueSource{
					GcpSecretRef: &GcpSecretRef{
						ProjectId: "test-project",
						Name:      "test-secret",
						Version:   "latest",
					},
				},
			},
			data:          &data,
			want:          "secret value",
			secretContent: "secret value",
			wantErr:       false,
		},
		{
			name: "with PayloadValueRef",
			propVal: PropertyAndValue{
				Value: &ignoreVal,
				PayloadValue: &PayloadValueRef{
					PropertyPaths: []string{"data.test2", "data.test"},
				},
			},
			data:          &data,
			want:          "data test value",
			secretContent: "data test value",
			wantErr:       false,
		},
		{
			name: "with PayloadValueRef with not existing path",
			propVal: PropertyAndValue{
				Value: &ignoreVal,
				PayloadValue: &PayloadValueRef{
					PropertyPaths: []string{"data.test2", "data.test3"},
				},
			},
			data:    &data,
			want:    "",
			wantErr: true,
		},
		{
			name: "ValueFrom and GcpSecretRef with path",
			propVal: PropertyAndValue{
				Value: &testVal,
				ValueFrom: &PropertyValueSource{
					GcpSecretRef: &GcpSecretRef{
						ProjectId: "test-project",
						Name:      "test-secret-json",
						Version:   "latest",
						Path:      "password",
					},
				},
			},
			data:          &data,
			want:          "test123",
			secretContent: `{"password":"test123"}`,
			wantErr:       false,
		},
		{
			name: "ValueFrom and GcpSecretRef with invalid path",
			propVal: PropertyAndValue{
				Value: &testVal,
				ValueFrom: &PropertyValueSource{
					GcpSecretRef: &GcpSecretRef{
						ProjectId: "test-project",
						Name:      "test-secret-json",
						Version:   "latest",
						Path:      "doesnotexist",
					},
				},
			},
			data:          &data,
			want:          "",
			secretContent: `{"password":"test123"}`,
			wantErr:       true,
		},
		{
			name: "From file",
			propVal: PropertyAndValue{

				FromFile: &fromFilePath,
			},
			data:    &data,
			want:    "from a file",
			wantErr: false,
		},
		{
			name: "From environment",
			propVal: PropertyAndValue{

				FromEnv: &testEnvName,
			},
			data:     &data,
			want:     "from env",
			envValue: "from env",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Mock server for GCP Secret Manager API
			if tt.propVal.ValueFrom != nil && tt.propVal.ValueFrom.GcpSecretRef != nil {
				testServer, client := gcp.NewFakeServerAndClient(ctx, t)
				secretName := fmt.Sprintf("projects/%s/secrets/%s/versions/%s", tt.propVal.ValueFrom.GcpSecretRef.ProjectId, tt.propVal.ValueFrom.GcpSecretRef.Name, tt.propVal.ValueFrom.GcpSecretRef.Version)
				crc32c := crc32.MakeTable(crc32.Castagnoli)
				checksum := int64(crc32.Checksum([]byte(tt.secretContent), crc32c))

				testServer.Responses[secretName] = gcp.FakeSecretManagerServerResponse{
					Response: &secretmanagerpb.AccessSecretVersionResponse{
						Name: secretName,
						Payload: &secretmanagerpb.SecretPayload{
							Data:       []byte(tt.secretContent),
							DataCrc32C: &checksum,
						},
					},
					Err: nil,
				}

				ctx = gcp.WithCtx(ctx, client)
				// gcp.SecretManagerBasePath = server.URL // Override the base path of the Secret Manager API
			}
			if tt.envValue != "" {
				os.Setenv(testEnvName, tt.envValue)
			}

			got, err := tt.propVal.GetValue(ctx, testLogger, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValue() = %v, want %v", got, tt.want)
			}
			// if got != tt.want {
			// 	t.Errorf("GetValue() = %v, want %v", got, tt.want)
			// }
		})
	}
}

func TestPropertyAndValue_GetStringValue(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	ctx := context.Background()

	type fields struct {
		propVal PropertyAndValue
	}
	type args struct {
		data *message.EventData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test with string value",
			fields: fields{
				propVal: PropertyAndValue{
					Value: "test value",
				},
			},
			args: args{
				data: &message.EventData{},
			},
			want:    "test value",
			wantErr: false,
		},
		{
			name: "Test with non-string value",
			fields: fields{
				propVal: PropertyAndValue{
					Value: 123,
				},
			},
			args: args{
				data: &message.EventData{},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.propVal.GetStringValue(ctx, testLogger, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("PropertyAndValue.GetStringValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("PropertyAndValue.GetStringValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPropertyAndValue_GetStringArrayValue(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	ctx := context.Background()

	type fields struct {
		propVal PropertyAndValue
	}
	type args struct {
		data *message.EventData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Test with string array value",
			fields: fields{
				propVal: PropertyAndValue{
					Value: []string{"test value1", "test value2"},
				},
			},
			args: args{
				data: &message.EventData{},
			},
			want:    []string{"test value1", "test value2"},
			wantErr: false,
		},
		{
			name: "Test with non-string array value",
			fields: fields{
				propVal: PropertyAndValue{
					Value: []int{123, 456},
				},
			},
			args: args{
				data: &message.EventData{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.propVal.GetStringArrayValue(ctx, testLogger, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("PropertyAndValue.GetStringArrayValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PropertyAndValue.GetStringArrayValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPropertyAndValue_GetMapStringStringValue(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	ctx := context.Background()

	type fields struct {
		propVal PropertyAndValue
	}
	type args struct {
		data *message.EventData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "Test with map string value",
			fields: fields{
				propVal: PropertyAndValue{
					Value: map[string]string{"key1": "value1", "key2": "value2"},
				},
			},
			args: args{
				data: &message.EventData{},
			},
			want:    map[string]string{"key1": "value1", "key2": "value2"},
			wantErr: false,
		},
		{
			name: "Test with non-map string value",
			fields: fields{
				propVal: PropertyAndValue{
					Value: "non-map string value",
				},
			},
			args: args{
				data: &message.EventData{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.propVal.GetMapStringStringValue(ctx, testLogger, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("PropertyAndValue.GetMapStringValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PropertyAndValue.GetMapStringValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPropertyAndValue_GetMapStringInterfaceValue(t *testing.T) {
	testLogger := zaptest.NewLogger(t)
	ctx := context.Background()

	type fields struct {
		propVal PropertyAndValue
	}
	type args struct {
		data *message.EventData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "Test with map string interface value",
			fields: fields{
				propVal: PropertyAndValue{
					Value: map[string]interface{}{"key1": "value1", "key2": "value2"},
				},
			},
			args: args{
				data: &message.EventData{},
			},
			want:    map[string]interface{}{"key1": "value1", "key2": "value2"},
			wantErr: false,
		},
		{
			name: "Test with map string value",
			fields: fields{
				propVal: PropertyAndValue{
					Value: map[string]string{"key1": "value1", "key2": "value2"},
				},
			},
			args: args{
				data: &message.EventData{},
			},
			want:    map[string]interface{}{"key1": "value1", "key2": "value2"},
			wantErr: false,
		},
		{
			name: "Test with non-map value",
			fields: fields{
				propVal: PropertyAndValue{
					Value: "non-map value",
				},
			},
			args: args{
				data: &message.EventData{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.propVal.GetMapStringInterfaceValue(ctx, testLogger, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("PropertyAndValue.GetMapStringInterfaceValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PropertyAndValue.GetMapStringInterfaceValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
