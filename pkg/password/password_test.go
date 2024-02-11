package password

import (
	"fmt"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	type args struct {
		length              int
		useLowerLetters     bool
		useUpperLetters     bool
		useSpecial          bool
		useNum              bool
		specialCharOverride string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Test with useLetters",
			args: args{
				length:              10,
				useLowerLetters:     true,
				useUpperLetters:     false,
				useSpecial:          false,
				useNum:              false,
				specialCharOverride: "",
			},
			want: 10,
		},
		{
			name: "Test with useSpecial",
			args: args{
				length:              10,
				useLowerLetters:     false,
				useUpperLetters:     false,
				useSpecial:          true,
				useNum:              false,
				specialCharOverride: "",
			},
			want: 10,
		},
		{
			name: "Test with useNum",
			args: args{
				length:              10,
				useLowerLetters:     false,
				useUpperLetters:     false,
				useSpecial:          false,
				useNum:              true,
				specialCharOverride: "",
			},
			want: 10,
		},
		{
			name: "Test with specialCharOverride",
			args: args{
				length:              10,
				useLowerLetters:     false,
				useUpperLetters:     false,
				useSpecial:          true,
				useNum:              false,
				specialCharOverride: "@!#$%^&*",
			},
			want: 10,
		},
		{
			name: "Test with everything enabled",
			args: args{
				length:          10,
				useLowerLetters: true,
				useUpperLetters: true,
				useSpecial:      true,
				useNum:          true,
			},
			want: 10,
		},
		// Add more test cases as needed
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GeneratePassword(tt.args.length, tt.args.useLowerLetters, tt.args.useUpperLetters, tt.args.useSpecial, tt.args.useNum, tt.args.specialCharOverride)
			fmt.Println(got)
			if len(got) != tt.want {
				t.Errorf("GeneratePassword() = %v, want %v", len(got), tt.want)
			}
		})
	}
}
