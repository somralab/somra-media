package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type memoryStore struct {
	mu     sync.Mutex
	nextID int64
	rows   map[int64]InstanceRecord
}

func newMemoryStore() *memoryStore {
	return &memoryStore{rows: make(map[int64]InstanceRecord)}
}

func (s *memoryStore) Create(_ context.Context, rec InstanceRecord) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, row := range s.rows {
		if row.PluginType == rec.PluginType && row.Name == rec.Name && rec.Name != "" {
			return 0, fmt.Errorf("create plugin instance: %w", ErrDuplicateInstance)
		}
	}
	s.nextID++
	id := s.nextID
	rec.ID = id
	if len(rec.Config) == 0 {
		rec.Config = json.RawMessage("{}")
	}
	s.rows[id] = rec
	return id, nil
}

func (s *memoryStore) UpdateConfig(_ context.Context, id int64, config json.RawMessage, secretsEnc string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.rows[id]
	if !ok {
		return ErrPluginNotFound
	}
	if len(config) == 0 {
		config = json.RawMessage("{}")
	}
	rec.Config = config
	rec.SecretsEnc = secretsEnc
	s.rows[id] = rec
	return nil
}

func (s *memoryStore) UpdateName(_ context.Context, id int64, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.rows[id]
	if !ok {
		return ErrPluginNotFound
	}
	for _, row := range s.rows {
		if row.ID != id && row.PluginType == rec.PluginType && row.Name == name && name != "" {
			return ErrDuplicateInstance
		}
	}
	rec.Name = name
	s.rows[id] = rec
	return nil
}

func (s *memoryStore) SetEnabled(_ context.Context, id int64, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.rows[id]
	if !ok {
		return ErrPluginNotFound
	}
	rec.Enabled = enabled
	s.rows[id] = rec
	return nil
}

func (s *memoryStore) GetByID(_ context.Context, id int64) (InstanceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.rows[id]
	if !ok {
		return InstanceRecord{}, ErrPluginNotFound
	}
	return rec, nil
}

func (s *memoryStore) List(_ context.Context) ([]InstanceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]InstanceRecord, 0, len(s.rows))
	for _, rec := range s.rows {
		out = append(out, rec)
	}
	return out, nil
}

func (s *memoryStore) Delete(_ context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.rows[id]; !ok {
		return ErrPluginNotFound
	}
	delete(s.rows, id)
	return nil
}
