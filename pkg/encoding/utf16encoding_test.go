package encoding

import "testing"

func TestEncodeStringToUtf16(t *testing.T) {
	testCases := []struct {
		name string
		data string
		want string
	}{
		{
			name: "Empty string",
			data: "",
			want: "",
		},
		{
			name: "ASCII string",
			data: "Hello, World!",
			want: "SABlAGwAbABvACwAIABXAG8AcgBsAGQAIQA=",
		},
		{
			name: "Unicode string",
			data: "Hello, 世界!",
			want: "SABlAGwAbABvACwAIAAWTkx1IQA=",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := EncodeStringToUtf16(tc.data)
			if got != tc.want {
				t.Errorf("EncodeStringToUtf16() = %v, want %v", got, tc.want)
			}
		})
	}
}
