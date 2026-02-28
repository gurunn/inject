# inject-lib

Общая библиотека для использования в разных приложениях: логгер, подключение к Sentry, хранилище (BadgerDB).

## Публикация на GitHub или GitLab

1. **Создайте отдельный репозиторий** (корень репозитория = эта папка `lib`).
   - GitHub: `github.com/<user>/inject-lib`
   - GitLab: `gitlab.com/<group>/inject-lib`

2. **Замените путь модуля в `go.mod`** на ваш:
   ```go
   module gitlab.com/yourgroup/inject-lib   // или github.com/youruser/inject-lib
   ```

3. Скопируйте в корень нового репо содержимое этой папки: `go.mod`, `go.sum`, `logger/`, `sentryconnect/`, `storage/`.

4. Создайте тег и запушьте:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```

## Импорт в приложении

```bash
go get github.com/gurunn/inject-lib@v0.1.0
# или
go get gitlab.com/yourgroup/inject-lib@v0.1.0
```

В коде:

```go
import (
    "github.com/gurunn/inject-lib/logger"
    "github.com/gurunn/inject-lib/sentryconnect"
    "github.com/gurunn/inject-lib/storage"
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

## Приватный репозиторий (GitHub/GitLab)

Импорт из приватного репо возможен. Нужно:

1. **Сказать Go не использовать публичный прокси** для вашего модуля:
   ```bash
   go env -w GOPRIVATE=github.com/youruser/inject-lib
   ```
   (или `gitlab.com/yourgroup/inject-lib`)

2. **Настроить доступ к GitHub/GitLab** (один из способов):
   - **HTTPS**: создать Personal Access Token с правом `repo` и настроить git, например:
     ```bash
     git config --global url."https://YOUR_TOKEN@github.com/".insteadOf "https://github.com/"
     ```
   - **SSH**: если ключи уже привязаны к аккаунту:
     ```bash
     git config --global url."git@github.com:".insteadOf "https://github.com/"
     ```

После этого `go get github.com/youruser/inject-lib@v0.1.0` будет работать с приватным репозиторием.

## Локальная разработка (до публикации)

В `go.mod` вашего приложения:

```go
require github.com/gurunn/inject-lib v0.0.0

replace github.com/gurunn/inject-lib => ./lib
```

После публикации репозитория уберите `replace` и укажите нужную версию.
