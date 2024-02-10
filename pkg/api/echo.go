package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kcloutie/event-reactor/pkg/http"
	"go.uber.org/zap"
)

func Echo(ctx context.Context, c *gin.Context) {
	var log *zap.Logger
	log, _ = http.SetCommonLoggingAttributes(ctx, c)
	slog := log.Sugar()

	slog.Debug("Echo request made", zap.String("path", c.Request.URL.Path), zap.String("remoteAddr", c.Request.RemoteAddr))
	attr := make(map[string]interface{})

	parts := strings.Split(c.Request.RemoteAddr, ":")
	attr["tcp"] = map[string]string{
		"ip":   strings.Join(parts[:(len(parts)-1)], ":"),
		"port": parts[len(parts)-1],
	}
	if c.Request.TLS != nil {
		attr["tls"] = map[string]string{
			"sni":    c.Request.TLS.ServerName,
			"cipher": tls.CipherSuiteName(c.Request.TLS.CipherSuite),
		}
	}
	headers := make(map[string]string)
	var cookies []string
	var buf bytes.Buffer
	if err := c.Request.Write(&buf); err != nil {
		slog.Errorf("Error reading request: %s", err)
		return
	}
	for name, value := range c.Request.Header {
		headers[name] = strings.Join(value, " ")
	}
	for _, cookie := range c.Request.Cookies() {
		cookies = append(cookies, cookie.String())
	}
	attr["http"] = map[string]interface{}{
		"protocol": c.Request.Proto,
		"headers":  headers,
		"cookies":  cookies,
		"host":     c.Request.Host,
		"method":   c.Request.Method,
		"path":     c.Request.URL.Path,
		"query":    c.Request.URL.RawQuery,
		"raw":      buf.String(),
	}
	res, _ := json.MarshalIndent(attr, "", "  ")
	log.Debug("Echo Response", zap.String("response", string(res)))
	c.Data(200, "application/json", res)
}
