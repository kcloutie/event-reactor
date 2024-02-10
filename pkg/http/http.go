package http

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kcloutie/event-reactor/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	RequestHeaderKey = "X-Request-Id"
)

var (
	TraceHeaderKey = "X-TRACE"
)

// https://www.baeldung.com/rest-api-error-handling-best-practices
type ErrorDetail struct {
	Type     string `json:"type,omitempty" yaml:"type,omitempty"`
	Title    string `json:"title,omitempty" yaml:"title,omitempty"`
	Status   int64  `json:"status,omitempty" yaml:"status,omitempty"`
	Detail   string `json:"detail,omitempty" yaml:"detail,omitempty"`
	Instance string `json:"instance,omitempty" yaml:"instance,omitempty"`
}

func SetCommonLoggingAttributes(ctx context.Context, c *gin.Context) (*zap.Logger, context.Context) {
	fields := []zapcore.Field{}

	fields = append(fields, zap.String("remoteIp", c.RemoteIP()))
	fields = append(fields, zap.String("clientIp", c.ClientIP()))
	fields = append(fields, zap.String("method", c.Request.Method))
	fields = append(fields, zap.Any("url", c.Request.URL))

	rid, _ := c.Get(RequestHeaderKey)
	fields = append(fields, zap.Any("requestId", rid.(string)))
	traceId, exists := c.Get(TraceHeaderKey)
	if exists {
		fields = append(fields, zap.Any("traceId", traceId.(string)))
	}

	return logger.FromCtxWithCtx(ctx, fields...)
}
