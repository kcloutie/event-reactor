package cel

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var desc = cel.Declarations(
	decls.NewVar("data", decls.NewMapType(decls.String, decls.Dyn)),
	decls.NewVar("attributes", decls.NewMapType(decls.String, decls.Dyn)),
)

func TestCelValue(t *testing.T) {
	type args struct {
		query string
		data  map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "string value",
			args: args{
				query: "data.prop1",

				data: map[string]interface{}{
					"data": map[string]interface{}{
						"prop1": "value1",
					},
				},
			},
			want: "value1",
		},
		{
			name: "bool value",
			args: args{
				query: "data.prop1",
				data: map[string]interface{}{
					"data": map[string]interface{}{
						"prop1": true,
					},
				},
			},
			want: "true",
		},
		{
			name: "array value",
			args: args{
				query: "data.prop1",
				data: map[string]interface{}{
					"data": map[string]interface{}{
						"prop1": []string{"value1", "value2"},
					},
				},
			},
			want: `["value1","value2"]`,
		},
		{
			name: "object value",
			args: args{
				query: "data.prop1",
				data: map[string]interface{}{
					"data": map[string]interface{}{
						"prop1": map[string]interface{}{
							"prop2": map[string]interface{}{
								"prop2": "dude",
							},
						},
					},
				},
			},
			want: `{"prop2":{"prop2":"dude"}}`,
		},
		{
			name: "bytes value",
			args: args{
				query: "data.prop1",
				data: map[string]interface{}{
					"data": map[string]interface{}{
						"prop1": []byte("value1"),
					},
				},
			},
			want: `"dmFsdWUx"`,
		},
		{
			name: "int value",
			args: args{
				query: "data.prop1",
				data: map[string]interface{}{
					"data": map[string]interface{}{
						"prop1": 1234,
					},
				},
			},
			want: "1234",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRaw, err := CelValue(tt.args.query, desc, tt.args.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("CelValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := strings.ReplaceAll(GetCelValue(gotRaw), " ", "")
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CelValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCelEvaluate(t *testing.T) {
	type args struct {
		ctx  context.Context
		expr string

		data map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    ref.Val
		wantErr bool
	}{
		{
			name: "simple boolean",
			args: args{
				ctx:  context.Background(),
				expr: "data.continue",
				data: map[string]interface{}{
					"data": map[string]interface{}{
						"continue": true,
					},
				},
			},
			want: types.True,
		},
		{
			name: "match deep",
			args: args{
				ctx:  context.Background(),
				expr: `data.child1.child2.prop.matches('this')`,
				data: map[string]interface{}{
					"data": map[string]interface{}{
						"child1": map[string]interface{}{
							"child2": map[string]interface{}{
								"prop": "please match this value",
							},
						},
					},
				},
			},
			want: types.True,
		},

		{
			name: "does not match deep",
			args: args{
				ctx:  context.Background(),
				expr: `data.child1.child2.prop.matches('nope')`,
				data: map[string]interface{}{
					"data": map[string]interface{}{
						"child1": map[string]interface{}{
							"child2": map[string]interface{}{
								"prop": "please match this value",
							},
						},
					},
				},
			},
			want: types.False,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CelEvaluate(tt.args.ctx, tt.args.expr, desc, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CelEvaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CelEvaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}
