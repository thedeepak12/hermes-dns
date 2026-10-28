package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type QueryLog struct {
	Timestamp    time.Time `json:"timestamp"`
	ClientIP     string    `json:"client_ip"`
	Domain       string    `json:"domain"`
	QueryType    string    `json:"query_type"`
	Result       string    `json:"result"`
	ResponseCode int       `json:"response_code"`
	LatencyMs    float64   `json:"latency_ms"`
}

type Logger struct {
	file    *os.File
	encoder *json.Encoder
	mu      sync.Mutex
	enabled bool
}

func New(logPath string, enabled bool) (*Logger, error) {
	if !enabled {
		return &Logger{enabled: false}, nil
	}

	dir := filepath.Dir(logPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return nil, err
    }

	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:    file,
		encoder: json.NewEncoder(file),
		enabled: true,
	}, nil
}

func (l *Logger) Log(query *QueryLog) {
	if !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.encoder.Encode(query)
}

func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}
