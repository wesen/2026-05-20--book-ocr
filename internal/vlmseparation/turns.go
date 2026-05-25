package vlmseparation

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

type TurnStore struct {
	store  chatstore.TurnStore
	convID string
	dsn    string
}

func OpenTurnStore(ctx context.Context, turnsDSN, turnsDB, outDir, convID string) (*TurnStore, func(), error) {
	_ = ctx
	if strings.TrimSpace(turnsDSN) == "" && strings.TrimSpace(turnsDB) == "" {
		turnsDB = filepath.Join(outDir, "turns.db")
	}
	dsn := strings.TrimSpace(turnsDSN)
	if dsn == "" {
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
	ts := &TurnStore{store: store, convID: convID, dsn: dsn}
	return ts, func() { _ = store.Close() }, nil
}

func (s *TurnStore) Save(ctx context.Context, sessionID, turnID, phase string, t *turns.Turn) error {
	if s == nil || s.store == nil || t == nil {
		return nil
	}
	if strings.TrimSpace(sessionID) == "" || strings.TrimSpace(turnID) == "" {
		return fmt.Errorf("turn store save requires sessionID and turnID")
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

func (s *TurnStore) DSN() string {
	if s == nil {
		return ""
	}
	return s.dsn
}
