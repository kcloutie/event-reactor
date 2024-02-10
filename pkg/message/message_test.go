package message

import (
	"reflect"
	"testing"
)

func TestEventData_AsMap(t *testing.T) {
	tests := []struct {
		name string
		n    EventData
		want map[string]interface{}
	}{
		{
			name: "All fields filled",
			n: EventData{
				Data: map[string]interface{}{
					"test": "123",
				},
				ID: "1",
				Attributes: map[string]string{
					"att1": "att1Val",
				},
			},
			want: map[string]interface{}{
				"data": map[string]interface{}{
					"test": "123",
				},
				"id": "1",
				"attributes": map[string]string{
					"att1": "att1Val",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.n.AsMap()

			if !reflect.DeepEqual(got["data"], tt.want["data"]) {
				t.Errorf("EventData.AsMap() data = %v, want %v", got["data"], tt.want["data"])
			}
			if !reflect.DeepEqual(got["id"], tt.want["id"]) {
				t.Errorf("EventData.AsMap() id = %v, want %v", got["id"], tt.want["id"])
			}

			if !reflect.DeepEqual(got["attributes"], tt.want["attributes"]) {
				t.Errorf("EventData.AsMap() attributes = %v, want %v", got["attributes"], tt.want["attributes"])
			}
		})
	}
}

func TestEventData_AsCelData(t *testing.T) {

	tests := []struct {
		name string
		n    EventData
		want map[string]interface{}
	}{
		{
			name: "basic",
			n: EventData{
				Data: map[string]interface{}{
					"test": "123",
				},
				ID: "1",
				Attributes: map[string]string{
					"att1": "att1Val",
				},
			},
			want: map[string]interface{}{
				"data": map[string]interface{}{
					"test": "123",
				},
				"id": "1",
				"attributes": map[string]string{
					"att1": "att1Val",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.AsMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EventData.AsCelData() = %v, want %v", got, tt.want)
			}
		})
	}
}
