package template

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"
)

func Test_PrefixSuffixStringArray(t *testing.T) {

	type args struct {
		prefix string
		suffix string
		users  []interface{}
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "prefix",
			args: args{
				prefix: "user:",
				users: []interface{}{
					"kcloutie@test.com",
					"abaker9@test.com",
				},
			},
			want: []string{
				"user:kcloutie@test.com",
				"user:abaker9@test.com",
			},
		},
		{
			name: "suffix",
			args: args{
				prefix: "",
				suffix: "@test.com",
				users: []interface{}{
					"kcloutie",
					"abaker9",
				},
			},
			want: []string{
				"kcloutie@test.com",
				"abaker9@test.com",
			},
		},
		{
			name: "prefix and suffix",
			args: args{
				prefix: "user:",
				suffix: "@test.com",
				users: []interface{}{
					"kcloutie",
					"abaker9",
				},
			},
			want: []string{
				"user:kcloutie@test.com",
				"user:abaker9@test.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrefixSuffixStringArray(tt.args.prefix, tt.args.suffix, tt.args.users)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toTerraformArrayWithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToHcl(t *testing.T) {
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "string",
			args: args{
				obj: "string_test",
			},
			want: "\"string_test\"",
		},
		{
			name: "int",
			args: args{
				obj: 1,
			},
			want: "1",
		},
		{
			name: "string array",
			args: args{
				obj: []string{"foo", "bar"},
			},
			want: "[\"foo\", \"bar\"]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPtr := ObjectToHcl(tt.args.obj)
			got := *gotPtr
			if got != tt.want {
				t.Errorf("ToHcl() = %v, want %v", &got, tt.want)
			}
		})
	}
}

func TestRemoveFromStringArray(t *testing.T) {
	type args struct {
		users    []interface{}
		toRemove string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "basic",
			args: args{
				users: []interface{}{
					"kcloutie@test.com",
					"abaker9@test.com",
				},
				toRemove: "@test.com",
			},
			want: []string{
				"kcloutie",
				"abaker9",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveFromStringArray(tt.args.toRemove, tt.args.users); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveFromStringArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrefixStringArray(t *testing.T) {
	type args struct {
		users  []interface{}
		prefix string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "basic",
			args: args{
				prefix: "user:",
				users: []interface{}{
					"kcloutie@test.com",
					"abaker9@test.com",
				},
			},
			want: []string{
				"user:kcloutie@test.com",
				"user:abaker9@test.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrefixStringArray(tt.args.prefix, tt.args.users); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PrefixStringArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSuffixStringArray(t *testing.T) {
	type args struct {
		users  []interface{}
		suffix string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{

		{
			name: "basic",
			args: args{

				suffix: "@test.com",
				users: []interface{}{
					"kcloutie",
					"abaker9",
				},
			},
			want: []string{
				"kcloutie@test.com",
				"abaker9@test.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SuffixStringArray(tt.args.suffix, tt.args.users); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SuffixStringArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJsonToHcl(t *testing.T) {
	type args struct {
		jsonString string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "array",
			args: args{
				jsonString: `{"test":["1", "2","3", "4"]}`,
			},
			want: `{ test = ["1", "2", "3", "4"] }`,
		},
		{
			name: "bool",
			args: args{
				jsonString: `{"test":true}`,
			},
			want: `{ test = true }`,
		},
		{
			name: "number",
			args: args{
				jsonString: `{"test":300}`,
			},
			want: `{ test = 300 }`,
		},
		{
			name: "complicated object",
			args: args{
				jsonString: `{"test":[{"array":["1", "2","3", "4"]},{"childObject": {"bool": true,"int":123, "string": "rrsss", "array":["1", "2","3", "4"], "decimal": 1.1 }}]}`,
			},
			want: `{ test = [{ array = ["1", "2", "3", "4"] }, { childObject = { array = ["1", "2", "3", "4"], bool = true, decimal = 1.1, int = 123, string = "rrsss" } }] }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JsonToHcl(tt.args.jsonString)
			if (err != nil) != tt.wantErr {
				t.Errorf("JsonToHcl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(*got, tt.want) {
				t.Errorf("JsonToHcl() = %v, want %v", *got, tt.want)
			}
		})
	}
}

// func TestWhere(t *testing.T) {
// 	kclustBytes, err := os.ReadFile("testdata/kubeClusterSource.json")
// 	if err != nil {
// 		t.Errorf("ArrayItemValue() - failed to read kubeClusterSource json file %v", err)
// 		return
// 	}

// 	var kclust []map[string]interface{}
// 	err = json.Unmarshal(kclustBytes, &kclust)
// 	if err != nil {
// 		t.Errorf("ArrayItemValue() - failed to unmarshal kubeClusterSource json file %v", err)
// 		return
// 	}
// 	type args struct {
// 		propertyName  string
// 		propertyValue string
// 		data          []map[string]interface{}
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want interface{}
// 	}{
// 		{
// 			name: "basic",
// 			args: args{
// 				propertyName:  "clusterName",
// 				propertyValue: "sb102",
// 				data:          kclust,
// 			},
// 			want: kclust[1],
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := Where(tt.args.propertyName, tt.args.propertyValue, tt.args.data); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Where() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestSelect(t *testing.T) {
// 	kclustBytes, err := os.ReadFile("testdata/kubeClusterSource.json")
// 	if err != nil {
// 		t.Errorf("ArrayItemValue() - failed to read kubeClusterSource json file %v", err)
// 		return
// 	}

// 	var kclust []map[string]interface{}
// 	err = json.Unmarshal(kclustBytes, &kclust)
// 	if err != nil {
// 		t.Errorf("ArrayItemValue() - failed to unmarshal kubeClusterSource json file %v", err)
// 		return
// 	}

// 	type args struct {
// 		propertyName string
// 		data         map[string]interface{}
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want interface{}
// 	}{
// 		{
// 			name: "basic",
// 			args: args{
// 				propertyName: "infrastructureName",
// 				data:         kclust[1],
// 			},
// 			want: "sb102-w5xtl",
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := Select(tt.args.propertyName, tt.args.data); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Select() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestWhere(t *testing.T) {

	kclustBytes, err := os.ReadFile("testdata/kubeClusterSource.json")
	if err != nil {
		t.Errorf("ArrayItemValue() - failed to read kubeClusterSource json file %v", err)
		return
	}

	var kclust []map[string]interface{}
	err = json.Unmarshal(kclustBytes, &kclust)
	if err != nil {
		t.Errorf("ArrayItemValue() - failed to unmarshal kubeClusterSource json file %v", err)
		return
	}
	type args struct {
		propertyName  string
		propertyValue string
		data          interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "array of map of interface",
			args: args{
				propertyName:  "clusterName",
				propertyValue: "sb102",
				data:          kclust,
			},
			want: kclust[1],
		},
		{
			name: "array of interface",
			args: args{
				propertyName:  "clusterName",
				propertyValue: "sb102",
				data: []interface{}{
					kclust[0],
					kclust[1],
				},
			},
			want: kclust[1],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Where(tt.args.propertyName, tt.args.propertyValue, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Where() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Where() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelect(t *testing.T) {

	kclustBytes, err := os.ReadFile("testdata/kubeClusterSource.json")
	if err != nil {
		t.Errorf("ArrayItemValue() - failed to read kubeClusterSource json file %v", err)
		return
	}

	var kclust []map[string]interface{}
	err = json.Unmarshal(kclustBytes, &kclust)
	if err != nil {
		t.Errorf("ArrayItemValue() - failed to unmarshal kubeClusterSource json file %v", err)
		return
	}
	type args struct {
		propertyName string
		data         interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				propertyName: "infrastructureName",
				data:         kclust[1],
			},
			want: "sb102-w5xtl",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Select(tt.args.propertyName, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Select() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertInterfaceToStringArray(t *testing.T) {
	// This may random error because the arrays items are in different order
	type args struct {
		data     any
		maxWidth int
	}
	tests := []struct {
		name string
		args args
		want [][][]string
	}{
		{
			name: "array of maps",
			args: args{
				maxWidth: 60,
				data: []map[string]interface{}{
					{
						"I1-row1-col1": "I1-row1-col2",
						"I1-row2-col1": "I1-row2-col2",
					},
					{
						"I2-row1-col1": "I2-row1-col2",
						"I2-row2-col1": "I2-row2-col2",
					},
				},
			},
			want: [][][]string{
				{
					{"I1-row1-col1", "I1-row1-col2"},
					{"I1-row2-col1", "I1-row2-col2"},
				},
				{
					{"I2-row1-col1", "I2-row1-col2"},
					{"I2-row2-col1", "I2-row2-col2"},
				},
			},
		},
		{
			name: "map",
			args: args{
				maxWidth: 60,
				data: map[string]interface{}{
					"I1-row1-col1": "I1-row1-col2",
					"I1-row2-col1": "I1-row2-col2",
				},
			},
			want: [][][]string{
				{
					{"I1-row1-col1", "I1-row1-col2"},
					{"I1-row2-col1", "I1-row2-col2"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ConvertInterfaceToStringArray(tt.args.data, tt.args.maxWidth)

			sort.Strings(got[0][0])
			sort.Strings(got[0][1])
			if len(got) > 1 {
				sort.Strings(got[1][0])
				sort.Strings(got[1][1])
			}
			sort.Strings(tt.want[0][0])
			sort.Strings(tt.want[0][1])
			if len(tt.want) > 1 {
				sort.Strings(tt.want[1][0])
				sort.Strings(tt.want[1][1])
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConvertInterfaceToStringArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToDnsString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "underscore",
			args: args{
				s: "1234_test",
			},
			want: "1234-test",
		},
		{
			name: "begin dash",
			args: args{
				s: "-1234test",
			},
			want: "1234test",
		},
		{
			name: "end dash",
			args: args{
				s: "1234test-",
			},
			want: "1234test",
		},
		{
			name: "space",
			args: args{
				s: "1234 test",
			},
			want: "1234-test",
		},
		{
			name: "over 256 chars",
			args: args{
				s: "1234testqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
			},
			want: "1234testqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToDnsString(tt.args.s); got != tt.want {
				t.Errorf("ToDnsString() = %v, want %v", got, tt.want)
			}
		})
	}
}
