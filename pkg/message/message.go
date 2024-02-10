package message

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	lcel "github.com/kcloutie/event-reactor/pkg/cel"
)

type EventData struct {
	Data       map[string]interface{}
	Attributes map[string]string
	ID         string
}

func (n EventData) AsMap() map[string]interface{} {
	return map[string]interface{}{
		"data":       n.Data,
		"attributes": n.Attributes,
		"id":         n.ID,
	}
}

func GetCelDecl() cel.EnvOption {
	return cel.Declarations(
		decls.NewVar("data", decls.NewMapType(decls.String, decls.Dyn)),
		decls.NewVar("attributes", decls.NewMapType(decls.String, decls.String)),
		decls.NewVar("id", decls.String),
	)
}

func (o *EventData) GetPropertyValue(property string) (string, error) {
	value, err := lcel.CelValue(property, GetCelDecl(), o.AsMap())
	if err != nil {
		return "", err
	}
	return lcel.GetCelValue(value), nil
}

func GenericPayloadToEventData(message interface{}) (EventData, error) {
	switch d := message.(type) {
	case map[string]interface{}:
		return EventData{
			Attributes: GetAttributes(d),
			ID:         "",
			Data:       d,
		}, nil
	case []map[string]interface{}:
		return EventData{
			Attributes: map[string]string{},
			ID:         "",
			Data: map[string]interface{}{
				"items": d,
			},
		}, nil
	case []interface{}:
		return EventData{
			Attributes: map[string]string{},
			ID:         "",
			Data: map[string]interface{}{
				"items": d,
			},
		}, nil

	default:
		return EventData{}, fmt.Errorf("failed to convert generic payload to EventData. Unknown type: %T", message)
	}

}

func GetAttributes(data map[string]interface{}) map[string]string {
	att, ok := data["attributes"]
	if !ok {
		return map[string]string{}
	}
	switch a := att.(type) {
	case map[string]interface{}:
		attributes := map[string]string{}
		for k, v := range a {
			attributes[k] = fmt.Sprintf("%v", v)
		}
	case map[string]string:
		return a
	}
	return map[string]string{}
}
