package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/kcloutie/event-reactor/pkg/config"
	httper "github.com/kcloutie/event-reactor/pkg/http"
	"github.com/kcloutie/event-reactor/pkg/listener/generic"
	"github.com/kcloutie/event-reactor/pkg/listener/pubsub"
	uuid "github.com/satori/go.uuid"
)

func CreateRouter(ctx context.Context, cacheInSeconds int) *gin.Engine {
	// log := logger.FromCtx(ctx)
	// slog := log.Sugar()
	router := gin.Default()

	router.Use(RequestIdMiddleware())
	router.Use(TraceLogsMiddleware())
	memoryStore := persist.NewMemoryStore(time.Duration(cacheInSeconds) * time.Second)

	router.GET("", cache.CacheByRequestURI(memoryStore, time.Duration(cacheInSeconds)*time.Second), func(c *gin.Context) {
		Home(ctx, c)
	})

	router.GET("/healthz", Health)
	router.GET("/readyz", Health)
	router.Any("/echo", func(c *gin.Context) {
		Echo(ctx, c)
	})

	apiV1 := router.Group("/api/v1")
	apiV1.Use()
	{
		phl := pubsub.New()
		apiV1.POST(fmt.Sprintf("/%s", phl.GetApiPath()), func(c *gin.Context) {
			ExecuteListener(ctx, c, phl)
		})

		genl := generic.New()
		apiV1.POST(fmt.Sprintf("/%s", genl.GetApiPath()), func(c *gin.Context) {
			ExecuteListener(ctx, c, genl)
		})

		// for _, l := range listener.GetListeners() {
		// 	err := l.Initialize(ctx)
		// 	if err != nil {
		// 		slog.Errorf("Failed to initialize listener %s. Error: %v", l.GetName(), err)
		// 		continue
		// 	}
		// 	slog.Debugf("Adding listener %s to router", l.GetName())
		// 	apiV1.POST(fmt.Sprintf("/%s", l.GetApiPath()), func(c *gin.Context) {
		// 		ExecuteListener(ctx, c, l)
		// 	})
		// }
	}
	return router
}

func Start(ctx context.Context, router *gin.Engine, cfg *config.ServerConfiguration, listeningAddr string) error {
	httper.TraceHeaderKey = cfg.TraceHeaderKey

	server := &http.Server{
		Addr:              listeningAddr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           router,
	}

	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func RequestIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := uuid.NewV4().String()
		c.Set(httper.RequestHeaderKey, rid)
		c.Writer.Header().Set(httper.RequestHeaderKey, rid)
		c.Next()
	}
}

func TraceLogsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Request.Header[httper.TraceHeaderKey]
		if exists {
			if len(val) == 1 {
				c.Set(httper.TraceHeaderKey, val[0])
			}
		}
		c.Next()
	}
}
