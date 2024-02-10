package message

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
)

func PubSubMessageToEventData(message map[string]interface{}) (EventData, error) {

	results := EventData{
		Attributes: map[string]string{},
	}

	root, exists := message["message"]
	if !exists {
		return EventData{}, fmt.Errorf("message property not found in the pub/sub message")
	}

	message, ok := root.(map[string]interface{})
	if !ok {
		return EventData{}, fmt.Errorf("message property is not of type map[string]interface{}")
	}

	attributes, exists := message["attributes"]

	if exists {
		switch a := attributes.(type) {
		case map[string]interface{}:
			for k, v := range a {
				results.Attributes[k] = fmt.Sprintf("%v", v)
			}
		case map[string]string:
			for k, v := range a {
				results.Attributes[k] = v
			}
		}
	}

	results.ID, _ = message["messageId"].(string)

	mesData := message["data"]

	data := map[string]interface{}{}

	switch d := mesData.(type) {
	case string:
		err := json.Unmarshal([]byte(d), &data)
		if err != nil {
			sDec, err := b64.StdEncoding.DecodeString(d)
			if err == nil {
				err = json.Unmarshal(sDec, &data)
				if err != nil {
					return results, fmt.Errorf("failed to unmarshal the pub/sub data property of the message (base64) - %v\nPAYLOAD:\n%s", err, string(d))
				}
			} else {
				return results, fmt.Errorf("failed to unmarshal the pub/sub data property of the message - %v\nPAYLOAD:\n%s", err, d)
			}

		}
	case []byte:
		err := json.Unmarshal(d, &data)
		if err != nil {
			return results, fmt.Errorf("failed to unmarshal the pub/sub data property of the message - %v\nPAYLOAD:\n%s", err, d)
		}
	}

	results.Data = data
	return results, nil
}
