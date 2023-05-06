package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"codeberg.org/cordalace/telepod/internal/podruntime"
	"codeberg.org/cordalace/telepod/internal/telegramnotifier"
	"codeberg.org/cordalace/telepod/internal/versionsdb"
	"codeberg.org/cordalace/telepod/internal/workflow"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var version = "dev"

type missingRequiredEnvError struct {
	Name string
}

func (e missingRequiredEnvError) Error() string {
	return fmt.Sprintf("missing required environment variable: %s", e.Name)
}

func requireEnv(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", missingRequiredEnvError{Name: name}
	}

	return value, nil
}

func env(name string, defaultValue string) string {
	var ret string

	value, ok := os.LookupEnv(name)
	if ok {
		ret = value
	} else {
		ret = defaultValue
	}

	return ret
}

var errInvalidLogLevel = errors.New("invalid log level")

func logLevelFromEnv(levelEnv string) (zapcore.Level, error) {
	var level zapcore.Level
	var err error

	switch levelEnv {
	case "DEBUG":
		level = zapcore.DebugLevel
	case "INFO":
		level = zapcore.InfoLevel
	case "WARN":
		level = zapcore.WarnLevel
	case "ERROR":
		level = zapcore.ErrorLevel
	default:
		err = errInvalidLogLevel
	}

	return level, err
}

func initLogger() (*zap.Logger, error) {
	levelEnv := env("LOG_LEVEL", "ERROR")

	logLevel, err := logLevelFromEnv(levelEnv)
	if err != nil {
		return nil, err
	}

	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level.SetLevel(logLevel)

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("error building logger: %w", err)
	}

	return logger, nil
}

func main() {
	logger, err := initLogger()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	logger.Info("telepod", zap.String("version", version))

	telegramChatID, err := requireEnv("TELEGRAM_CHAT_ID")
	if err != nil {
		logger.Fatal("error loading env var", zap.Error(err))
	}
	telegramBotToken, err := requireEnv("TELEGRAM_BOT_TOKEN")
	if err != nil {
		logger.Fatal("error loading env var", zap.Error(err))
	}

	podRuntime := podruntime.NewPodRuntime()
	if err := podRuntime.Init(); err != nil {
		logger.Fatal("error initializing pod runtime", zap.Error(err))
	}

	versionsDB := versionsdb.NewVersionsDB()
	if err := versionsDB.Init(); err != nil {
		logger.Fatal("error initializing versions db", zap.Error(err))
	}

	httpClient := new(http.Client)

	telegramNotifier := telegramnotifier.NewTelegramNotifier(httpClient, telegramChatID, telegramBotToken)

	wf := workflow.NewWorkflow(podRuntime, versionsDB, telegramNotifier)

	if err := wf.Run(context.Background()); err != nil {
		logger.Fatal("error running workflow", zap.Error(err))
	}
}
