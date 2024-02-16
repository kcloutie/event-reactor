package api

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kcloutie/event-reactor/pkg/adapter"
	"github.com/kcloutie/event-reactor/pkg/config"
	"github.com/kcloutie/event-reactor/pkg/http"
	"github.com/kcloutie/event-reactor/pkg/listener"
	"github.com/kcloutie/event-reactor/pkg/matcher"
	"github.com/kcloutie/event-reactor/pkg/message"
	"github.com/kcloutie/event-reactor/pkg/reactor"

	"go.uber.org/zap"
)

func ExecuteListener(ctx context.Context, c *gin.Context, listener listener.ListenerInterface) {
	cfg := config.FromCtx(ctx)
	var log *zap.Logger
	log, ctx = http.SetCommonLoggingAttributes(ctx, c)
	slog := log.Sugar()

	slog.Debugf("Executing listener '%s'", listener.GetName())

	if c.Request.Body == nil {
		errorMes := "request body was empty, request cannot be processed"
		errD := &http.ErrorDetail{
			Type:     listener.GetName() + "-get-request-body",
			Title:    listener.GetName() + " Get Request Body",
			Status:   400,
			Detail:   errorMes,
			Instance: listener.GetApiPath(),
		}
		log.Error(errorMes)
		WriteResponse(slog, int(errD.Status), []http.ErrorDetail{*errD}, c, cfg)
		return
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		errD := &http.ErrorDetail{
			Type:     listener.GetName() + "-get-request-body",
			Title:    listener.GetName() + " Get Request Body",
			Status:   400,
			Detail:   err.Error(),
			Instance: listener.GetApiPath(),
		}
		log.Error(err.Error())
		WriteResponse(slog, int(errD.Status), []http.ErrorDetail{*errD}, c, cfg)
		return
	}
	if cfg.LogRawPubSubPayload {
		log.Info("raw Pub/Sub Payload", zap.String("payload", string(payload)))
	}
	eventPayload, errD := listener.ParsePayload(ctx, log, payload)
	if errD != nil {
		log.Error(errD.Detail)

		WriteResponse(slog, int(errD.Status), []http.ErrorDetail{*errD}, c, cfg)
		return
	}

	log = log.With(zap.String("message_id", eventPayload.ID))

	if cfg.LogEventDataPayload {
		log.Info("eventPayload Payload", zap.Any("eventPayload", eventPayload))
	}

	reactorFunctions := adapter.GetReactorNewFunctions(cfg.LoadTestReactor)

	if len(cfg.ReactorConfigs) == 0 {
		slog.Warnf("no reactors configured for listener '%s'", listener.GetName())
	}

	errors := RunReactorsAsync(ctx, cfg, log, eventPayload, listener.GetName(), listener.GetApiPath(), reactorFunctions)
	if len(errors) > 0 {
		WriteResponse(slog, 400, errors, c, cfg)
		return
	}
}

func RunReactorsAsync(ctx context.Context, cfg *config.ServerConfiguration, log *zap.Logger, eventPayload *message.EventData, listenerName string, listenerApiPath string, reactorFunctions map[string]func(log *zap.Logger, reactorConfig config.ReactorConfig) reactor.ReactorInterface) []http.ErrorDetail {
	channels := []chan http.ErrorDetail{}
	errors := []http.ErrorDetail{}
	wg := new(sync.WaitGroup)

	for i, reactorConfig := range cfg.ReactorConfigs {
		wg.Add(1)
		ch := make(chan http.ErrorDetail, 1)
		channels = append(channels, ch)
		defer close(ch)
		log := log.With(zap.String("reactorName", reactorConfig.Name), zap.String("reactorType", reactorConfig.Type))
		go executeReactors(wg, channels[i], ctx, reactorConfig, eventPayload, listenerName, listenerApiPath, log, reactorFunctions)
	}
	wg.Wait()

	for _, ch := range channels {
		select {
		case errD := <-ch:
			errors = append(errors, errD)
		default:
		}
	}
	return errors
}

func WriteResponse(log *zap.SugaredLogger, status int, errD []http.ErrorDetail, c *gin.Context, cfg *config.ServerConfiguration) {
	if len(errD) > 0 && cfg.AlwaysReturn200 {
		log.Warn("at least one error occurred however the server is configured to always return 200...returning 200")
		status = 200
	}

	c.JSON(status, errD)
}

func executeReactors(wg *sync.WaitGroup, ch chan http.ErrorDetail, ctx context.Context, reactorConfig config.ReactorConfig, eventPayload *message.EventData, listenerName string, listenerApiPath string, log *zap.Logger, reactorFunctions map[string]func(log *zap.Logger, reactorConfig config.ReactorConfig) reactor.ReactorInterface) {
	matches, err := matcher.Matches(ctx, reactorConfig, eventPayload)
	if err != nil {
		errD := http.ErrorDetail{
			Type:     listenerName + "-match-message",
			Title:    listenerName + "Match Message",
			Status:   400,
			Detail:   err.Error(),
			Instance: listenerApiPath,
		}
		log.Error(errD.Detail)

		ch <- errD
		wg.Done()
		return
	}
	if !matches {
		// log.Debug(fmt.Sprintf("reactor '%s' of type '%s' does not match message", reactorConfig.Name, reactorConfig.Type))
		wg.Done()
		return
	}

	newReactorFunc, exists := reactorFunctions[reactorConfig.Type]
	if !exists {
		errD := http.ErrorDetail{
			Type:     listenerName + "-exists",
			Title:    listenerName + " Exists",
			Status:   400,
			Detail:   fmt.Sprintf("reactor type of '%s' does not exist. Verify the reactor type of '%s' within the configuration", reactorConfig.Type, reactorConfig.Name),
			Instance: listenerApiPath,
		}
		log.Error(errD.Detail)

		ch <- errD
		wg.Done()
		return
	}

	reactorObj := newReactorFunc(log, reactorConfig)

	log.Debug(fmt.Sprintf("executing reactor '%s' of type '%s'", reactorConfig.Name, reactorObj.GetName()))
	err = reactorObj.ProcessEvent(ctx, eventPayload)
	if err != nil {
		errD := http.ErrorDetail{
			Type:     listenerName + "-" + reactorObj.GetName() + "-execute-reactor",
			Title:    listenerName + "-" + reactorObj.GetName() + " Execute Reactor",
			Status:   400,
			Detail:   err.Error(),
			Instance: listenerApiPath,
		}
		log.Error(errD.Detail)

		if !reactorConfig.GetFailOnError() {
			wg.Done()
			return
		}
		ch <- errD
		wg.Done()
		return
	}

	log.Debug(fmt.Sprintf("execution of reactor '%s' of type '%s' has completed successfully", reactorConfig.Name, reactorObj.GetName()))
	wg.Done()
}
