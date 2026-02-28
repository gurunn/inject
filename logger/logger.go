package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/getsentry/sentry-go"
	log "github.com/sirupsen/logrus"
)

type BorderFormatter struct {
	Base *log.TextFormatter
}

func (f *BorderFormatter) Format(entry *log.Entry) ([]byte, error) {
	message, err := f.Base.Format(entry)
	if err != nil {
		return nil, err
	}

	// Добавляем рамку ТОЛЬКО для ошибок
	if entry.Level == log.ErrorLevel {
		border := "+----------------------------------------+"
		return []byte(fmt.Sprintf("%s\n%s%s\n", border, message, border)), nil
	}

	return message, nil
}

func LogError(err error, functionName string, additionalFields ...map[string]interface{}) {
	if err == nil {
		return
	}

	fields := log.Fields{
		"error":    err.Error(),
		"function": functionName,
	}

	// Добавляем информацию о месте вызова
	if pc, file, line, ok := runtime.Caller(1); ok {
		fields["file"] = fmt.Sprintf("%s: %d", filepath.Base(file), line)
		fields["func"] = runtime.FuncForPC(pc).Name()
	}

	// Добавляем дополнительные поля
	if len(additionalFields) > 0 {
		for k, v := range additionalFields[0] {
			fields[k] = v
		}
	}

	log.WithFields(fields).Error()
}

// InitLogger инициализирует логгер с учетом окружения
func InitLogger() {
	// Установка уровня логирования
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.SetLevel(log.InfoLevel)
		log.Errorf("Invalid LOG_LEVEL '%s', defaulting to INFO", logLevel)
	} else {
		log.SetLevel(level)
	}

	// Конфигурация форматера
	isK8s := os.Getenv("KUBERNETES_SERVICE_HOST") != ""

	base := &log.TextFormatter{
		DisableQuote:    true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
		ForceColors:     !isK8s,
	}
	formatter := &BorderFormatter{Base: base}

	// В Kubernetes всегда пишем в stdout
	if isK8s {
		log.SetOutput(os.Stdout)
		formatter.Base.TimestampFormat = "2006-01-02T15:04:05.000Z07:00"
	}

	log.SetFormatter(formatter)
}

// LogAndCapture объединяет логирование и отправку в Sentry
func LogAndCapture(hub *sentry.Hub, err error, context string, additionalFields ...map[string]interface{}) {
	LogError(err, context, additionalFields...)

	if hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetExtra("context", context)
			hub.CaptureException(err)
		})
	}
}

// CleanupLogger очищает буферы и хукеры логгера
func CleanupLogger() {
	log.StandardLogger().Hooks = make(log.LevelHooks)

	if formatter, ok := log.StandardLogger().Formatter.(*BorderFormatter); ok {
		if formatter.Base != nil {
			formatter.Base.DisableQuote = true
			formatter.Base.FullTimestamp = true
			formatter.Base.TimestampFormat = "2006-01-02 15:04:05.000"
			formatter.Base.ForceColors = !isK8sEnvironment()
		}
	}
}

func isK8sEnvironment() bool {
	return os.Getenv("KUBERNETES_SERVICE_HOST") != ""
}
