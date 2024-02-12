package matcher

import (
	"context"
	"fmt"

	"github.com/google/cel-go/common/types"
	"github.com/kcloutie/event-reactor/pkg/cel"
	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/logger"
	"github.com/kcloutie/event-reactor/pkg/message"
	"go.uber.org/zap"
)

func Matches(ctx context.Context, reactorConfig config.ReactorConfig, data *message.EventData) (bool, error) {
	log := logger.FromCtx(ctx).With(zap.String("reactorName", reactorConfig.Name), zap.String("reactorType", reactorConfig.Type)).Sugar()
	if reactorConfig.Disabled {
		log.Debugf("reactor '%s' is disabled, skipping event", reactorConfig.Name)
		return false, nil

	}
	if reactorConfig.CelExpressionFilter == "" {
		log.Debugf("reactor '%s' has no CEL expression filter so it matches all events", reactorConfig.Name)
		return true, nil
	}

	matches, err := cel.CelEvaluate(ctx, reactorConfig.CelExpressionFilter, message.GetCelDecl(), data.AsMap())
	if err != nil {
		log.Error(fmt.Sprintf("error evaluating CEL expression on reactor '%s'", reactorConfig.Name), zap.Error(err))
		return false, nil
	}
	if matches == types.True {

		log.Debugf("message matched reactor '%s' CEL filtering", reactorConfig.Name)
		return true, nil
	}
	log.Debugf("message did not match reactor '%s' CEL filtering", reactorConfig.Name)
	return false, nil

}
