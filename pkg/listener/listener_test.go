package listener

import (
	"sort"
	"testing"
)

func TestGetListeners(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "basic",
			want: []string{"generic", "pubsub"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetListeners()

			sort.Slice(got, func(i, j int) bool {
				return got[i].GetApiPath() < got[j].GetApiPath()
			})

			sort.Strings(tt.want)

			for i := 0; i < len(tt.want); i++ {
				if tt.want[i] != got[i].GetApiPath() {
					t.Errorf("GetListeners() = %v, want %v", got[i].GetApiPath(), tt.want[i])
				}
			}

		})
	}
}
