package cel

import (
	"context"
	"encoding/json"
	"fmt"

	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	structType = reflect.TypeOf(&structpb.Value{})
	listType   = reflect.TypeOf(&structpb.ListValue{})
	mapType    = reflect.TypeOf(&structpb.Struct{})
)

func celEvaluateValue(expr string, env *cel.Env, data map[string]interface{}) (ref.Val, error) {
	parsed, issues := env.Parse(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("failed to parse expression %#v: %w", expr, issues.Err())
	}

	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("expression %#v check failed: %w", expr, issues.Err())
	}

	prg, err := env.Program(checked, cel.EvalOptions(cel.OptOptimize))
	if err != nil {
		return nil, fmt.Errorf("expression %#v failed to create a Program: %w", expr, err)
	}

	out, _, err := prg.Eval(data)
	if err != nil {
		return nil, fmt.Errorf("expression %#v failed to evaluate: %w", expr, err)
	}

	return out, nil
}

func CelValue(query string, declarations cel.EnvOption, data map[string]interface{}) (ref.Val, error) {

	celDec, _ := cel.NewEnv(declarations)
	val, err := celEvaluateValue(query, celDec, data)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func CelEvaluate(ctx context.Context, expr string, declarations cel.EnvOption, data map[string]interface{}) (ref.Val, error) {

	env, err := cel.NewEnv(declarations)

	if err != nil {
		return nil, err
	}

	parsed, issues := env.Parse(expr)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("failed to parse expression %#v: %w", expr, issues.Err())
	}

	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("expression %#v check failed: %w", expr, issues.Err())
	}

	prg, err := env.Program(checked)
	if err != nil {
		return nil, fmt.Errorf("expression %#v failed to create a Program: %w", expr, err)
	}

	out, _, err := prg.Eval(data)
	if err != nil {
		return nil, fmt.Errorf("expression %#v failed to evaluate: %w", expr, err)
	}
	return out, nil
}

func GetCelValue(val ref.Val) string {
	var raw interface{}
	var b []byte
	var err error

	switch val.(type) {
	case types.String:
		if v, ok := val.Value().(string); ok {
			b = []byte(v)
		}
	case types.Bytes:
		raw, err = val.ConvertToNative(structType)
		if err == nil {
			b, err = raw.(*structpb.Value).MarshalJSON()
			if err != nil {
				b = []byte{}
			}
		}
	case types.Double, types.Int:
		raw, err = val.ConvertToNative(structType)
		if err == nil {
			b, err = raw.(*structpb.Value).MarshalJSON()
			if err != nil {
				b = []byte{}
			}
		}
	case traits.Lister:
		raw, err = val.ConvertToNative(listType)
		if err == nil {
			s, err := protojson.Marshal(raw.(proto.Message))
			if err == nil {
				b = s
			}
		}
	case traits.Mapper:
		raw, err = val.ConvertToNative(mapType)
		if err == nil {
			s, err := protojson.Marshal(raw.(proto.Message))
			if err == nil {
				b = s
			}
		}
	case types.Bool:
		raw, err = val.ConvertToNative(structType)
		if err == nil {
			b, err = json.Marshal(raw.(*structpb.Value).GetBoolValue())
			if err != nil {
				b = []byte{}
			}
		}

	default:
		raw, err = val.ConvertToNative(reflect.TypeOf([]byte{}))
		if err == nil {
			if v, ok := raw.([]byte); ok {
				b = v
			}
		}

	}
	return string(b)
}
