package matcher

import (
	"context"
	"testing"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/message"
)

func TestMatches(t *testing.T) {
	data := map[string]interface{}{
		"prop1": "val1",
		"prop2": map[string]interface{}{
			"childProp1": "childVal1",
		},
	}
	attributes := map[string]string{
		"att1": "val1",
	}
	type args struct {
		ctx           context.Context
		reactorConfig config.ReactorConfig
		data          *message.EventData
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "no filter",
			args: args{
				ctx: context.Background(),
				reactorConfig: config.ReactorConfig{
					Name:                "no filter",
					CelExpressionFilter: "",
					Disabled:            false,
				},
				data: &message.EventData{
					ID:         "1",
					Data:       data,
					Attributes: attributes,
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "disabled",
			args: args{
				ctx: context.Background(),
				reactorConfig: config.ReactorConfig{
					Name:                "disabled",
					CelExpressionFilter: "",
					Disabled:            true,
				},
				data: &message.EventData{
					ID:         "1",
					Data:       data,
					Attributes: attributes,
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "success CEL filter",
			args: args{
				ctx: context.Background(),
				reactorConfig: config.ReactorConfig{
					Name:                "with filter",
					CelExpressionFilter: "data.prop1 == 'val1'",
					Disabled:            false,
				},
				data: &message.EventData{
					ID:         "1",
					Data:       data,
					Attributes: attributes,
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "failed CEL filter",
			args: args{
				ctx: context.Background(),
				reactorConfig: config.ReactorConfig{
					Name:                "with filter",
					CelExpressionFilter: "data.prop1 == 'val2'",
					Disabled:            false,
				},
				data: &message.EventData{
					ID:         "1",
					Data:       data,
					Attributes: attributes,
				},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Matches(tt.args.ctx, tt.args.reactorConfig, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Matches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}
