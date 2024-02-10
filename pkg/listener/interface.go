package listener

import (
	"context"

	"github.com/kcloutie/event-reactor/pkg/http"
	"github.com/kcloutie/event-reactor/pkg/message"
	"go.uber.org/zap"
)

type ListenerInterface interface {
	Initialize(ctx context.Context) error
	GetName() string
	GetApiPath() string
	ParsePayload(ctx context.Context, log *zap.Logger, payload []byte) (*message.EventData, *http.ErrorDetail)
}
