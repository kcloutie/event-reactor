package matcher

import (
	"context"

	"github.com/google/cel-go/common/types"
	"github.com/kcloutie/event-reactor/pkg/cel"
	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/logger"
	"github.com/kcloutie/event-reactor/pkg/message"
	"go.uber.org/zap"
)

func Matches(ctx context.Context, reactorConfig config.ReactorConfig, data *message.EventData) (bool, error) {
	log := logger.FromCtx(ctx).With(zap.String("reactorConfig", reactorConfig.Name))
	if reactorConfig.Disabled {
		log.Debug("reactor config is disabled, skipping event")
		return false, nil

	}
	if reactorConfig.CelExpressionFilter == "" {
		log.Debug("reactor config has no CEL expression filter so it matches all events")
		return true, nil
	}

	matches, err := cel.CelEvaluate(ctx, reactorConfig.CelExpressionFilter, message.GetCelDecl(), data.AsMap())
	if err != nil {
		log.Error("error evaluating CEL expression", zap.Error(err))
		return false, nil
	}
	if matches == types.True {
		log.Debug("message matched reactor config CEL filtering")
		return true, nil
	}
	log.Debug("message did not match reactor config CEL filtering")
	return false, nil

}
