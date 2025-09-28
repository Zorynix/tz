// @title Subscription Service API
// @version 1.0
// @description REST API for managing online subscription data
// @host localhost:8080
// @BasePath /api/v1
package main

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "subscription-service/docs"
	appfx "subscription-service/internal/app/fx"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	zapLogger, _ := zap.NewDevelopment()
	defer func() {
		_ = zapLogger.Sync()
	}()

	app := fx.New(
		fx.WithLogger(func() fxevent.Logger {
			return &fxevent.ZapLogger{Logger: zapLogger}
		}),
		appfx.Module(),
	)

	app.Run()
}