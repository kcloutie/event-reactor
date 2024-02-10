package email

import (
	"context"
	"reflect"
	"testing"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/message"
	"github.com/kcloutie/event-reactor/pkg/reactor"
	"go.uber.org/zap"
)

func TestReactor_GetReactorConfig(t *testing.T) {
	ctx := context.Background()
	log, _ := zap.NewProduction()
	defer log.Sync()

	type fields struct {
		reactorConfig config.ReactorConfig
	}
	type args struct {
		data *message.EventData
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Test with valid properties",
			fields: fields{
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"smtpHost":   {Value: reactor.ToPtrString("smtp.example.com")},
						"smtpPort":   {Value: reactor.ToPtrString("587")},
						"maxRetries": {Value: reactor.ToPtrString("5")},
						"from":       {Value: reactor.ToPtrString("from@example.com")},
						"to":         {Value: reactor.ToPtrString("to@example.com")},
						"password":   {Value: reactor.ToPtrString("password")},
						"subject":    {Value: reactor.ToPtrString("subject")},
						"body":       {Value: reactor.ToPtrString("body")},
					},
				},
			},
			args: args{
				data: &message.EventData{},
			},
			wantErr: false,
		},
		{
			name: "Test with missing smtpHost",
			fields: fields{
				reactorConfig: config.ReactorConfig{
					Properties: map[string]config.PropertyAndValue{
						"smtpPort":   {Value: reactor.ToPtrString("587")},
						"maxRetries": {Value: reactor.ToPtrString("5")},
						"from":       {Value: reactor.ToPtrString("from@example.com")},
						"to":         {Value: reactor.ToPtrString("to@example.com")},
						"password":   {Value: reactor.ToPtrString("password")},
						"subject":    {Value: reactor.ToPtrString("subject")},
						"body":       {Value: reactor.ToPtrString("body")},
					},
				},
			},
			args: args{
				data: &message.EventData{},
			},
			wantErr: true,
		},
		// Add more test cases as needed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reactor{
				Log:           log,
				reactorConfig: tt.fields.reactorConfig,
			}
			_, err := r.GetReactorConfig(ctx, tt.args.data, log)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reactor.GetReactorConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestReactor_GetProperties(t *testing.T) {
	r := &Reactor{}
	want := []config.ReactorConfigProperty{
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
	if got := r.GetProperties(); !reflect.DeepEqual(got, want) {
		t.Errorf("Reactor.GetProperties() = %v, want %v", got, want)
	}
}
func Test_splitEmailAddress(t *testing.T) {
	type args struct {
		address string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Split by semicolon",
			args: args{
				address: "test1@test.com;test2@test.com;test3@test.com",
			},
			want: []string{"test1@test.com", "test2@test.com", "test3@test.com"},
		},
		{
			name: "Split by comma",
			args: args{
				address: "test1@test.com,test2@test.com,test3@test.com",
			},
			want: []string{"test1@test.com", "test2@test.com", "test3@test.com"},
		},
		{
			name: "Split by space",
			args: args{
				address: "test1@test.com test2@test.com test3@test.com",
			},
			want: []string{"test1@test.com", "test2@test.com", "test3@test.com"},
		},
		{
			name: "No delimiters",
			args: args{
				address: "test1@test.com",
			},
			want: []string{"test1@test.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitEmailAddress(tt.args.address); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitEmailAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
