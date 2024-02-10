package generic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kcloutie/event-reactor/pkg/http"
	"github.com/kcloutie/event-reactor/pkg/message"
	"go.uber.org/zap"
)

type Listener struct {
	Name    string
	ApiPath string
}

func New() *Listener {
	return &Listener{
		Name:    "generic",
		ApiPath: "generic",
	}
}

func (v *Listener) Initialize(ctx context.Context) error {
	return nil
}

func (v *Listener) GetName() string {
	return v.Name
}

func (v *Listener) GetApiPath() string {
	return v.ApiPath
}

func (v *Listener) ParsePayload(ctx context.Context, log *zap.Logger, payload []byte) (*message.EventData, *http.ErrorDetail) {
	var request interface{}

	err := json.Unmarshal(payload, &request)
	if err != nil {
		mess := fmt.Sprintf("Failed to unmarshal body to map[string]interface{} type. Error: %v", err)
		errD := &http.ErrorDetail{
			Type:     "unmarshal-body-data",
			Title:    "Unmarshal Body Data",
			Status:   400,
			Detail:   mess,
			Instance: v.GetApiPath(),
		}
		log.Error(mess)

		return nil, errD
	}

	notifyData, err := message.GenericPayloadToEventData(request)
	if err != nil {
		errD := &http.ErrorDetail{
			Type:     "convert-generic-payload",
			Title:    "Convert Generic Payload",
			Status:   400,
			Detail:   err.Error(),
			Instance: v.GetApiPath(),
		}
		log.Error(err.Error())
		return nil, errD
	}
	return &notifyData, nil
}
