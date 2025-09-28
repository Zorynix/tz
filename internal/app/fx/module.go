package fx

import (
	"context"
	"fmt"

	"subscription-service/internal/config"
	"subscription-service/internal/handler"
	"subscription-service/internal/repository"
	"subscription-service/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Options(
		LoggerComponent(),
		StorageComponent(),
		RepositoryComponent(),
		ServiceComponent(),
		HandlerComponent(),
		HTTPComponent(),

		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),
	)
}

func LoggerComponent() fx.Option {
	return fx.Provide(
		LoadConfig,
		NewLogger,
	)
}

func StorageComponent() fx.Option {
	return fx.Options(
		fx.Provide(NewDatabase),
		fx.Invoke(RegisterDatabaseLifecycle),
	)
}

func RepositoryComponent() fx.Option {
	return fx.Provide(NewSubscriptionRepository)
}

func ServiceComponent() fx.Option {
	return fx.Provide(NewSubscriptionService)
}

func HandlerComponent() fx.Option {
	return fx.Provide(NewSubscriptionHandler)
}

func HTTPComponent() fx.Option {
	return fx.Options(
		fx.Provide(NewGinServer),
		fx.Invoke(RegisterHTTPServer),
	)
}

func LoadConfig() (*config.Config, error) {
	configPath := "config.docker.yaml"
	return config.Load(configPath)
}

func NewLogger(cfg *config.Config) *zap.Logger {
	var zapConfig zap.Config

	if cfg.Logger.Level == "debug" {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	if cfg.Logger.Encoding == "console" {
		zapConfig.Encoding = "console"
	}

	logger, err := zapConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	return logger
}

func NewDatabase(logger *zap.Logger, cfg *config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	logger.Info("connecting to database",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("dbname", cfg.Database.DBName),
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Error("failed to create connection pool", zap.Error(err))
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		logger.Error("failed to ping database", zap.Error(err))
		return nil, err
	}

	logger.Info("database connection established")
	return pool, nil
}

func NewSubscriptionRepository(db *pgxpool.Pool, logger *zap.Logger) repository.SubscriptionRepository {
	return repository.NewSubscriptionRepository(db, logger)
}

func NewSubscriptionService(repo repository.SubscriptionRepository, logger *zap.Logger) service.SubscriptionService {
	return service.NewSubscriptionService(repo, logger)
}

func NewSubscriptionHandler(svc service.SubscriptionService, logger *zap.Logger) *handler.SubscriptionHandler {
	return handler.NewSubscriptionHandler(svc, logger)
}

func RegisterDatabaseLifecycle(lc fx.Lifecycle, logger *zap.Logger, db *pgxpool.Pool) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("closing database connection")
			db.Close()
			return nil
		},
	})
}
