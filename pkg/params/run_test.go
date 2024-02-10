package params

import "testing"

func TestStringToBool(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "true",
			args: args{
				s: "true",
			},
			want: true,
		},
		{
			name: "Yes",
			args: args{
				s: "Yes",
			},
			want: true,
		},
		{
			name: "1",
			args: args{
				s: "1",
			},
			want: true,
		},
		{
			name: "random char",
			args: args{
				s: "dude",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringToBool(tt.args.s); got != tt.want {
				t.Errorf("StringToBool() = %v, want %v", got, tt.want)
			}
		})
	}
}
