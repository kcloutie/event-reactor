package pubsub

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kcloutie/event-reactor/pkg/message"
	"go.uber.org/zap/zaptest"
)

func TestListener_ParsePayload(t *testing.T) {
	goodPayload := map[string]interface{}{
		"message": map[string]interface{}{
			"attributes": map[string]string{
				"att1": "att1Val",
			},
			"data": []byte(`{"test":"123"}`),
			"id":   "1",
		},
	}
	goodPayloadBytes, _ := json.Marshal(&goodPayload)
	expectedGood, _ := message.PubSubMessageToEventData(goodPayload)
	type args struct {
		payload []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *message.EventData
		wantErr string
	}{
		{
			name: "non json payload",
			args: args{
				payload: []byte(`dude`),
			},
			want:    nil,
			wantErr: "Failed to unmarshal body to the map[string]interface{} type. Error: invalid character 'd' looking for beginning of value",
		},
		{
			name: "bad payload",
			args: args{
				payload: []byte(`{}`),
			},
			want:    nil,
			wantErr: "message property not found in the pub/sub message",
		},
		{
			name: "success",
			args: args{
				payload: goodPayloadBytes,
			},
			want:    &expectedGood,
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLogger := zaptest.NewLogger(t)

			l := New()
			got, errD := l.ParsePayload(context.Background(), testLogger, tt.args.payload)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Listener.ParsePayload() got = %v, want %v", got, tt.want)
			}

			if errD == nil {
				if tt.wantErr != "" {
					t.Errorf("Listener.ParsePayload() err = nil, want %v", tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(errD.Detail, tt.wantErr) {
				t.Errorf("Listener.ParsePayload() err = %v, want %v", errD.Detail, tt.wantErr)
			}
		})
	}
}
