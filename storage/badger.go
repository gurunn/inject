package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/getsentry/sentry-go"
	"github.com/gurunn/inject/logger"
	log "github.com/sirupsen/logrus"
)

// Badger — хранилище на BadgerDB. AfterInsert вызывается асинхронно после Insert (опционально).
type Badger struct {
	db           *badger.DB
	Path         string
	SentryBadger *sentry.Hub
	// AfterInsert — опциональный коллбэк после успешной записи (например, логирование в БД приложения).
	AfterInsert func(taskID string, data []byte)
	closed      bool
	mutex       sync.RWMutex
	cancelGC    context.CancelFunc
}

func (b *Badger) Open() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.db != nil {
		return nil
	}

	opts := badger.DefaultOptions(b.Path)
	opts.ValueLogFileSize = 16 << 20
	opts.MemTableSize = 4 << 20
	opts.NumMemtables = 2
	opts.NumLevelZeroTables = 2
	opts.NumLevelZeroTablesStall = 3
	opts.CompactL0OnClose = true
	opts.ValueThreshold = 64 << 10

	var err error
	b.db, err = badger.Open(opts)
	if err != nil {
		logger.LogAndCapture(b.SentryBadger, err, "Failed to open badger database", map[string]interface{}{
			"path": b.Path,
		})
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	b.cancelGC = cancel
	go b.valueLogGCWorker(ctx)

	log.Info("BadgerDB opened with optimized settings")
	return nil
}

func (b *Badger) valueLogGCWorker(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.mutex.RLock()
			if b.closed || b.db == nil {
				b.mutex.RUnlock()
				return
			}
			b.mutex.RUnlock()

			err := b.db.RunValueLogGC(0.7)
			if err != nil && err != badger.ErrNoRewrite && err != badger.ErrRejected {
				log.Warnf("ValueLog GC failed: %v", err)
			}
		}
	}
}

func (b *Badger) Insert(taskID string, jsonData []byte) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if b.closed || b.db == nil {
		return errors.New("database is closed")
	}

	if len(jsonData) > 10*1024*1024 {
		return fmt.Errorf("data too large: %d bytes", len(jsonData))
	}

	err := b.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(taskID), jsonData).WithTTL(300 * time.Second)
		return txn.SetEntry(entry)
	})

	if err != nil {
		logger.LogAndCapture(b.SentryBadger, err, "Failed to insert task", map[string]interface{}{
			"task_id": taskID,
			"size":    len(jsonData),
		})
		return fmt.Errorf("failed to save result for task %s: %v", taskID, err)
	}

	if b.AfterInsert != nil {
		go func(id string, data []byte) {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("AfterInsert panic: %v", r)
				}
			}()
			if len(data) > 5*1024*1024 {
				log.Warnf("Large data skipped for async processing: %d bytes", len(data))
				return
			}
			b.AfterInsert(id, data)
		}(taskID, jsonData)
	}

	return nil
}

func (b *Badger) Get(taskID string) (*TaskResult, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if b.closed || b.db == nil {
		return nil, errors.New("database is closed")
	}

	var taskResult TaskResult
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(taskID))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			var data interface{}
			if err := json.Unmarshal(val, &data); err != nil {
				return err
			}
			taskResult = TaskResult{Status: "completed", Data: data}
			return nil
		})
	})

	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return &TaskResult{Status: "pending"}, nil
		}
		logger.LogAndCapture(b.SentryBadger, err, "Failed to get task", map[string]interface{}{
			"task_id": taskID,
		})
		return nil, err
	}

	return &taskResult, nil
}

func (b *Badger) Close() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.db == nil {
		return nil
	}

	b.closed = true
	if b.cancelGC != nil {
		b.cancelGC()
	}

	err := b.db.Close()
	b.db = nil

	if err != nil {
		logger.LogAndCapture(b.SentryBadger, err, "Failed to close BadgerDB", nil)
		return err
	}

	log.Info("BadgerDB closed successfully")
	return nil
}
