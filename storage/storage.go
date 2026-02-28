package storage

// TaskResult — результат задачи (pending/completed/failed). Используется в интерфейсе Database.
type TaskResult struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}

// Database — интерфейс хранилища задач.
type Database interface {
	Open() error
	Insert(taskID string, jsonData []byte) error
	Get(taskID string) (*TaskResult, error)
}
