package reactor

import (
	"context"

	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/message"
	"go.uber.org/zap"
)

type ReactorInterface interface {
	GetName() string
	GetDescription() string
	SetLogger(logger *zap.Logger)
	SetReactor(reactorConfig config.ReactorConfig)
	ProcessEvent(ctx context.Context, eventData *message.EventData) error
	GetHelp() string
	GetConfigExample() string
	GetProperties() []config.ReactorConfigProperty
	GetRequiredPropertyNames() []string
}
