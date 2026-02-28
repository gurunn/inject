# github.com/gurunn/inject

Общая библиотека для использования в разных приложениях: логгер, подключение к Sentry, хранилище (BadgerDB).

## Импорт в приложении

```bash
go get github.com/gurunn/inject@v0.3.0
```

В коде:

```go
import (
    "github.com/gurunn/inject/logger"
    "github.com/gurunn/inject/sentryconnect"
    "github.com/gurunn/inject/storage"
)

// Логгер
logger.InitLogger()
logger.LogError(err, "context")
logger.LogAndCapture(hub, err, "context", nil)

// Sentry (version и moduleName из вашего приложения)
hub, err := sentryconnect.InitSentry(version, moduleName)

// Хранилище
db := &storage.Badger{
    Path:         "/tmp/badger",
    SentryBadger: hub,
    AfterInsert:  func(taskID string, data []byte) { /* опционально */ },
}
db.Open()
db.Insert(taskID, jsonData)
res, err := db.Get(taskID) // *storage.TaskResult
```

## Checksum-база

Если Go обращается к checksum-базе и выдаёт ошибку:

```bash
go env -w GONOSUMDB=github.com/gurunn/inject
```

## Приватный репозиторий

1. Отключите прокси для модуля:
   ```bash
   go env -w GOPRIVATE=github.com/gurunn/inject
   ```

2. Настройте доступ (HTTPS с токеном или SSH), например:
   ```bash
   git config --global url."git@github.com:".insteadOf "https://github.com/"
   ```

## Локальная разработка

В `go.mod` вашего приложения:

```go
require github.com/gurunn/inject v0.0.0

replace github.com/gurunn/inject => /path/to/inject
```

После публикации уберите `replace` и укажите нужную версию (например, `v0.3.0`).
