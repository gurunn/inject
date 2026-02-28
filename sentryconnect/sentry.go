package sentryconnect

import (
	"fmt"
	"os"

	"github.com/getsentry/sentry-go"
)

// InitSentry инициализирует Sentry. version и moduleName передаются из приложения (напр. models.Version, models.NameRelation).
func InitSentry(version, moduleName string) (*sentry.Hub, error) {
	dsn := os.Getenv("SENTRY_DSN_GO")
	if dsn == "" {
		return nil, fmt.Errorf("SENTRY_DSN_GO environment variable not set")
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		TracesSampleRate: 1.0,
		AttachStacktrace: true,
		Release:          version,
		Debug:            os.Getenv("SENTRY_DEBUG") == "true",
	})
	if err != nil {
		return nil, err
	}

	hub := sentry.CurrentHub().Clone()
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("module", moduleName)
	})

	return hub, nil
}
