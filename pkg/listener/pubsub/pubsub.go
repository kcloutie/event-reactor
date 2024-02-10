package pubsub

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
		Name:    "pub/sub",
		ApiPath: "pubsub",
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

// https://cloud.google.com/pubsub/docs/publish-receive-messages-client-library
// https://cloud.google.com/pubsub/docs/samples/pubsub-subscribe-avro-records
func (v *Listener) ParsePayload(ctx context.Context, log *zap.Logger, payload []byte) (*message.EventData, *http.ErrorDetail) {
	var request map[string]interface{}
	err := json.Unmarshal(payload, &request)
	if err != nil {
		mess := fmt.Sprintf("Failed to unmarshal body to the map[string]interface{} type. Error: %v", err)
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

	log.Debug("Pub/Sub Payload", zap.Any("payload", payload))

	notifyData, err := message.PubSubMessageToEventData(request)
	if err != nil {
		errD := &http.ErrorDetail{
			Type:     "convert-pubsub-message",
			Title:    "Convert Pub/Sub Message",
			Status:   400,
			Detail:   err.Error(),
			Instance: v.GetApiPath(),
		}
		log.Error(err.Error())
		return nil, errD
	}
	return &notifyData, nil
}
