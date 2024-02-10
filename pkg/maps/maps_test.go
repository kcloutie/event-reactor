package maps

import (
	"encoding/json"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"
)

var (
	testFolderName      = "testdata"
	mainMapPath         = path.Join(testFolderName, "mainMap.json")
	mapToMergePath      = path.Join(testFolderName, "mapToMerge.json")
	expectWithMergePath = path.Join(testFolderName, "expectedWithMerge.json")
)

func TestGetStringFromMapByPath(t *testing.T) {
	type args struct {
		path   MapPath
		tknObj map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "string value exists",
			args: args{
				path: "level1.level2.stringProperty",
				tknObj: map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"stringProperty": "stringProperty",
						},
					},
				},
			},
			want: "stringProperty",
		},
		{
			name: "string value does not exist",
			args: args{
				path: "level1.level2.stringProperty2",
				tknObj: map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"stringProperty": "stringProperty",
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "middle value is not map",
			args: args{
				path: "level1.level2.level3",
				tknObj: map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": "not a map",
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValueFromMapByPath(tt.args.path, tt.args.tknObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStringFromMapByPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetStringFromMapByPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePath(t *testing.T) {
	type args struct {
		path MapPath
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "basic",
			args: args{
				path: NewMapPath("level1.level2.level3"),
			},
			want: []string{
				"level1",
				"level2",
				"level3",
			},
		},
		{
			name: "with periods in section and single quotes",
			args: args{
				path: NewMapPath("level1.level2.level3.'this.is.a.test'"),
			},
			want: []string{
				"level1",
				"level2",
				"level3",
				"this.is.a.test",
			},
		},
		{
			name: "with periods in section and single and double quotes",
			args: args{
				path: NewMapPath("level1.\"level.2\".level3.'this.is.a.test'"),
			},
			want: []string{
				"level1",
				"level.2",
				"level3",
				"this.is.a.test",
			},
		},
		{
			name: "real life failure",
			args: args{
				path: NewMapPath("metadata.annotations.'pipelinesascode.tekton.dev/pipeline'"),
			},
			want: []string{
				"metadata",
				"annotations",
				"pipelinesascode.tekton.dev/pipeline",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.path.ParsePath()

			// gotBytes, _ := json.MarshalIndent(got, "", "  ")
			// os.WriteFile("got.json", gotBytes, 0644)

			// wantBytes, _ := json.MarshalIndent(tt.want, "", "  ")
			// os.WriteFile("want.json", wantBytes, 0644)

			if len(got) != len(tt.want) {
				t.Errorf("ParsePath() = Len - %v, want %v", len(got), len(tt.want))
				return
			}

			for i := 0; i < len(got); i++ {
				if !reflect.DeepEqual(got[i], tt.want[i]) {
					t.Errorf("ParsePath() = %v, want %v", got, tt.want)
				}

			}

		})
	}
}

func TestMapPathPieces_ToMapPath(t *testing.T) {
	tests := []struct {
		name   string
		pieces MapPathPieces
		want   MapPath
	}{
		{
			name: "basic",
			pieces: MapPathPieces{
				"level1",
				"level2",
				"level3",
			},
			want: "level1.level2.level3",
		},
		{
			name: "with periods",
			pieces: MapPathPieces{
				"level1",
				"level2",
				"level3",
				"level.4",
			},
			want: "level1.level2.level3.'level.4'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pieces.ToMapPath(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapPathPieces.ToMapPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapPathPieces_RemoveFirstPiece(t *testing.T) {
	tests := []struct {
		name   string
		pieces MapPathPieces
		want   MapPath
	}{
		{
			name: "basic",
			pieces: MapPathPieces{
				"level1",
				"level2",
				"level3",
			},
			want: "level2.level3",
		},
		{
			name: "with periods",
			pieces: MapPathPieces{
				"level1",
				"level2",
				"level3",
				"level.4",
			},
			want: "level2.level3.'level.4'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pieces.RemoveFirstPiece(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapPathPieces.RemoveFirstPiece() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStringValueFromMapByPath(t *testing.T) {
	floatVal, _ := strconv.ParseFloat("448146263476", 64)
	type args struct {
		path             MapPath
		tknObj           map[string]interface{}
		throwIfNotString bool
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				path: "level1.level2.stringParam",
				tknObj: map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"stringParam": "stringValue",
						},
					},
				},
				throwIfNotString: true,
			},
			want: "stringValue",
		},
		{
			name: "int",
			args: args{
				path: "level1.level2.intVal",
				tknObj: map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"intVal": 448146263476,
						},
					},
				},
				throwIfNotString: true,
			},
			want: "448146263476",
		},
		{
			name: "float",
			args: args{
				path: "level1.level2.floatVal",
				tknObj: map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": map[string]interface{}{
							"floatVal": floatVal,
						},
					},
				},
				throwIfNotString: true,
			},
			want: "448146263476",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStringValueFromMapByPath(tt.args.path, tt.args.tknObj, tt.args.throwIfNotString)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStringValueFromMapByPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetStringValueFromMapByPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeMaps(t *testing.T) {
	mainMap := map[string]interface{}{}
	mainMapBytes, err := os.ReadFile(mainMapPath)
	if err != nil {
		t.Errorf("MergeMaps() - Unable to open the mappings file")
		return
	}
	json.Unmarshal(mainMapBytes, &mainMap)

	mapToMerge := map[string]interface{}{}
	mapToMergeBytes, err := os.ReadFile(mapToMergePath)
	if err != nil {
		t.Errorf("MergeMaps() - Unable to open the default mappings file")
		return
	}
	json.Unmarshal(mapToMergeBytes, &mapToMerge)

	expectWithMerge := map[string]interface{}{}
	expectWithMergeBytes, err := os.ReadFile(expectWithMergePath)
	if err != nil {
		t.Errorf("MergeMaps() - Unable to open the default mappings file")
		return
	}
	json.Unmarshal(expectWithMergeBytes, &expectWithMerge)
	type args struct {
		a map[string]interface{}
		b map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "1 level",
			args: args{
				a: map[string]interface{}{
					"key1": "val1",
					"key2": "val2",
				},
				b: map[string]interface{}{
					"key2": "changed",
					"key3": "val3",
				},
			},
			want: map[string]interface{}{
				"key1": "val1",
				"key2": "changed",
				"key3": "val3",
			},
		},

		{
			name: "2 levels",
			args: args{
				a: map[string]interface{}{
					"key1": "val1",
					"key2": "val2",
					"parent1": map[string]interface{}{
						"child1": "val1",
						"child2": "val2",
					},
				},
				b: map[string]interface{}{
					"key2": "changed",
					"key3": "val3",
					"parent1": map[string]interface{}{
						"child1": "changed",
						"child2": "val2",
						"child3": "val3",
					},
				},
			},
			want: map[string]interface{}{
				"key1": "val1",
				"key2": "changed",
				"key3": "val3",
				"parent1": map[string]interface{}{
					"child1": "changed",
					"child2": "val2",
					"child3": "val3",
				},
			},
		},
		{
			name: "default mappings and mappings",
			args: args{
				a: mainMap,
				b: mapToMerge,
			},
			want: expectWithMerge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeMaps(tt.args.a, tt.args.b)

			// gotBytes, _ := json.MarshalIndent(got, "", "  ")
			// os.WriteFile("got.json", gotBytes, 0644)

			// wantBytes, _ := json.MarshalIndent(tt.want, "", "  ")
			// os.WriteFile("want.json", wantBytes, 0644)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeMaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHaveSameKey(t *testing.T) {
	type args struct {
		a map[string]interface{}
		b map[string]interface{}
	}
	tests := []struct {
		name  string
		args  args
		want  bool
		want1 string
	}{
		{
			name: "has same key",
			args: args{
				a: map[string]interface{}{
					"key1": "val1",
					"key2": "val2",
					"key3": "val3",
				},
				b: map[string]interface{}{
					"key4": "val4",
					"key5": "val5",
					"key3": "val3",
				},
			},
			want:  true,
			want1: "key3",
		},
		{
			name: "does not have same key",
			args: args{
				a: map[string]interface{}{
					"key1": "val1",
					"key2": "val2",
					"key3": "val3",
				},
				b: map[string]interface{}{
					"key4": "val4",
					"key5": "val5",
					"key6": "val6",
				},
			},
			want:  false,
			want1: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := HaveSameKey(tt.args.a, tt.args.b)
			if got != tt.want {
				t.Errorf("HaveSameKey() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("HaveSameKey() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestMergeStringMaps(t *testing.T) {
	type args struct {
		a map[string]string
		b map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "basic",
			args: args{
				a: map[string]string{
					"key1": "val1",

					"key3": "val3",
				},
				b: map[string]string{

					"key2": "val2",
				},
			},
			want: map[string]string{
				"key1": "val1",
				"key3": "val3",
				"key2": "val2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeStringMaps(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeStringMaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyableSlice_DeepCopy(t *testing.T) {
	tests := []struct {
		name string
		s    CopyableSlice
		want []interface{}
	}{
		{
			name: "basic",
			s: CopyableSlice{
				"1",
				"2",
				3,
				[]interface{}{
					1, 2,
				},
				map[string]interface{}{
					"key1": "val1",
				},
			},
			want: []interface{}{
				"1",
				"2",
				3,
				[]interface{}{
					1, 2,
				},
				map[string]interface{}{
					"key1": "val1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.DeepCopy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CopyableSlice.DeepCopy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyableMap_DeepCopy(t *testing.T) {
	tests := []struct {
		name string
		m    CopyableMap
		want map[string]interface{}
	}{
		{
			name: "basic",
			m: CopyableMap{
				"key1": "val1",
				"key2": 2,
				"key3": []interface{}{
					1,
					2,
				},
				"key4": map[string]interface{}{
					"key1": "val1",
				},
			},
			want: map[string]interface{}{
				"key1": "val1",
				"key2": 2,
				"key3": []interface{}{
					1,
					2,
				},
				"key4": map[string]interface{}{
					"key1": "val1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.DeepCopy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CopyableMap.DeepCopy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeepCopyInterface(t *testing.T) {
	type args struct {
		mapObj interface{}
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
				mapObj: map[string]interface{}{
					"key1": "val1",
					"key2": 2,
					"key3": map[string]interface{}{
						"key1": "val1",
					},
				},
			},
			want: map[string]interface{}{
				"key1": "val1",
				"key2": float64(2),
				"key3": map[string]interface{}{
					"key1": "val1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeepCopyInterface(tt.args.mapObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeepCopyInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeepCopyInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeepCopyMapPointer(t *testing.T) {
	type args struct {
		mapObj *map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *map[string]interface{}
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				mapObj: &map[string]interface{}{
					"key1": "val1",
					"key2": 2,
					"key3": map[string]interface{}{
						"key1": "val1",
					},
				},
			},
			want: &map[string]interface{}{
				"key1": "val1",
				"key2": float64(2),
				"key3": map[string]interface{}{
					"key1": "val1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DeepCopyMapPointer(tt.args.mapObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeepCopyMapPointer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeepCopyMapPointer() = %v, want %v", got, tt.want)
			}
		})
	}
}
