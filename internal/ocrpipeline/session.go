package ocrpipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/geppetto/pkg/turns"
	"github.com/go-go-golems/geppetto/pkg/turns/serde"
	chatstore "github.com/go-go-golems/pinocchio/pkg/persistence/chatstore"
)

type TurnStoreConfig struct {
	TurnsDSN string
	TurnsDB  string
	WorkDir  string
	ConvID   string
}

type OCRTurnStore struct {
	store  chatstore.TurnStore
	convID string
	dsn    string
}

func OpenOCRTurnStore(cfg TurnStoreConfig) (*OCRTurnStore, func(), error) {
	convID := strings.TrimSpace(cfg.ConvID)
	if convID == "" {
		return nil, nil, fmt.Errorf("convID is required")
	}
	dsn := strings.TrimSpace(cfg.TurnsDSN)
	turnsDB := strings.TrimSpace(cfg.TurnsDB)
	if dsn == "" {
		if turnsDB == "" {
			if strings.TrimSpace(cfg.WorkDir) == "" {
				return nil, nil, fmt.Errorf("workDir or turnsDB is required when turnsDSN is empty")
			}
			turnsDB = filepath.Join(cfg.WorkDir, "turns.db")
		}
		if err := os.MkdirAll(filepath.Dir(turnsDB), 0o755); err != nil {
			return nil, nil, err
		}
		var err error
		dsn, err = chatstore.SQLiteTurnDSNForFile(turnsDB)
		if err != nil {
			return nil, nil, err
		}
	}
	store, err := chatstore.NewSQLiteTurnStore(dsn)
	if err != nil {
		return nil, nil, err
	}
	turnStore := &OCRTurnStore{store: store, convID: convID, dsn: dsn}
	return turnStore, func() { _ = store.Close() }, nil
}

func (s *OCRTurnStore) Save(ctx context.Context, sessionID, turnID, phase string, t *turns.Turn) error {
	if s == nil || s.store == nil || t == nil {
		return nil
	}
	if strings.TrimSpace(sessionID) == "" || strings.TrimSpace(turnID) == "" || strings.TrimSpace(phase) == "" {
		return fmt.Errorf("sessionID, turnID, and phase are required")
	}
	payload, err := serde.ToYAML(t, serde.Options{})
	if err != nil {
		return err
	}
	runtimeKey := ""
	if v, ok, err := turns.KeyTurnMetaRuntime.Get(t.Metadata); err == nil && ok {
		runtimeKey = fmt.Sprint(v)
	}
	inferenceID := ""
	if v, ok, err := turns.KeyTurnMetaInferenceID.Get(t.Metadata); err == nil && ok {
		inferenceID = v
	}
	return s.store.Save(ctx, s.convID, sessionID, turnID, phase, time.Now().UnixMilli(), string(payload), chatstore.TurnSaveOptions{RuntimeKey: runtimeKey, InferenceID: inferenceID})
}

func (s *OCRTurnStore) DSN() string {
	if s == nil {
		return ""
	}
	return s.dsn
}

func BookRunConvID(bookID, runID string) string {
	return fmt.Sprintf("book-ocr:%s:%s", bookID, runID)
}

func PageSessionID(page int) string {
	return fmt.Sprintf("page:%03d", page)
}

func PageTurnID(page int, index int, name string) string {
	return fmt.Sprintf("page:%03d:%02d-%s", page, index, name)
}
