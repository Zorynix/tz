package fx

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"subscription-service/internal/config"
	"subscription-service/internal/handler"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type contextKey string

const startTimeKey contextKey = "start_time"

func NewGinServer(subscriptionHandler *handler.SubscriptionHandler, logger *zap.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), startTimeKey, start))
		c.Next()
	})

	handler.SetupRoutes(router, subscriptionHandler, logger)

	logger.Info("gin server initialized")
	return router
}

func RegisterHTTPServer(
	lc fx.Lifecycle,
	router *gin.Engine,
	logger *zap.Logger,
	cfg *config.Config,
) {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting http server", zap.String("addr", server.Addr))

			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("failed to start http server", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping http server")
			return server.Shutdown(ctx)
		},
	})
}
