package message

import (
	"reflect"
	"testing"
)

func TestToToEventData(t *testing.T) {
	tests := []struct {
		name    string
		message map[string]interface{}
		want    EventData
		wantErr bool
	}{
		{
			name: "All fields filled",
			message: map[string]interface{}{
				"message": map[string]interface{}{
					"attributes": map[string]string{
						"att1": "att1Val",
					},
					"data":      []byte(`{"test":"123"}`),
					"messageId": "1",
				},
			},
			want: EventData{
				Attributes: map[string]string{
					"att1": "att1Val",
				},
				Data: map[string]interface{}{
					"test": "123",
				},
				ID: "1",
			},
			wantErr: false,
		},
		{
			name: "Some fields empty",
			message: map[string]interface{}{
				"message": map[string]interface{}{
					"attributes": map[string]string{},
					"data":       []byte(`{}`),
					"messageId":  "1",
				},
			},
			want: EventData{
				Attributes: map[string]string{},
				Data:       map[string]interface{}{},
				ID:         "1",
			},
			wantErr: false,
		},
		{
			name: "Invalid JSON data",
			message: map[string]interface{}{
				"message": map[string]interface{}{
					"attributes": map[string]string{
						"att1": "att1Val",
					},
					"data":      []byte(`dude`),
					"messageId": "1",
				},
			},
			want: EventData{
				Attributes: map[string]string{
					"att1": "att1Val",
				},
				ID: "1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PubSubMessageToEventData(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToEventData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToEventData() got = %v, want %v", got, tt.want)
			}
		})
	}
}
