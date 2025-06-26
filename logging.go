package main

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"
)

var Yellow = "\033[33m"
var Red = "\033[31m"
var Cyan = "\033[36m"
var Reset = "\033[0m"
var White = "\033[97m"

type JSONLogRecord struct {
	Time    time.Time      `json:"time"`
	Level   string         `json:"level"`
	Message string         `json:"message"`
	Attrs   map[string]any `json:"attrs,omitempty"`
}

type ApplicationLogHandler struct {
	mu      sync.Mutex
	records []JSONLogRecord
}

func (h *ApplicationLogHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *ApplicationLogHandler) Handle(_ context.Context, r slog.Record) error {
	logEntry := JSONLogRecord{
		Time:    r.Time,
		Level:   r.Level.String(),
		Message: r.Message,
		Attrs:   map[string]any{},
	}

	r.Attrs(func(attr slog.Attr) bool {
		logEntry.Attrs[attr.Key] = attr.Value.Any()
		return true
	})

	h.mu.Lock()
	h.records = append(h.records, logEntry)
	h.mu.Unlock()

	var levelColour string

	switch r.Level {
	case slog.LevelDebug:
		levelColour = Cyan
	case slog.LevelInfo:
		levelColour = White
	case slog.LevelWarn:
		levelColour = Yellow
	case slog.LevelError:
		levelColour = Red
	}

	fmt.Printf("[%s%s%s] %s %s\n", levelColour, r.Level.String(), Reset, r.Time.Format(time.RFC3339), r.Message)
	return nil
}

func (h *ApplicationLogHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

func (h *ApplicationLogHandler) WithGroup(_ string) slog.Handler {
	return h
}

func (h *ApplicationLogHandler) GetRecords() []JSONLogRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.records)
}
