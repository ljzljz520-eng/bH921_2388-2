package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"weddinglive/internal/domain"
)

type JSONStore struct {
	mu    sync.RWMutex
	path  string
	state domain.State
}

func NewMemory(initial domain.State) *JSONStore {
	return &JSONStore{state: normalize(initial.Clone())}
}

func Open(path string, initial domain.State) (*JSONStore, error) {
	if path == "" {
		return nil, errors.New("data file path is required")
	}

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		s := &JSONStore{path: path, state: normalize(initial.Clone())}
		if err := s.persist(s.state); err != nil {
			return nil, err
		}
		return s, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read data file: %w", err)
	}

	var state domain.State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode data file: %w", err)
	}
	return &JSONStore{path: path, state: normalize(state)}, nil
}

func (s *JSONStore) View(fn func(domain.State) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(s.state.Clone())
}

func (s *JSONStore) Update(fn func(*domain.State) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	working := s.state.Clone()
	if err := fn(&working); err != nil {
		return err
	}
	working = normalize(working)
	if s.path != "" {
		if err := s.persist(working); err != nil {
			return err
		}
	}
	s.state = working
	return nil
}

func (s *JSONStore) Snapshot() domain.State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state.Clone()
}

func (s *JSONStore) persist(state domain.State) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode data file: %w", err)
	}
	data = append(data, '\n')
	temporary := s.path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o600); err != nil {
		return fmt.Errorf("write data file: %w", err)
	}
	if err := os.Rename(temporary, s.path); err != nil {
		return fmt.Errorf("replace data file: %w", err)
	}
	return nil
}

func normalize(state domain.State) domain.State {
	if state.NextAccountID < 1 {
		state.NextAccountID = 1
	}
	if state.NextRoomID < 1 {
		state.NextRoomID = 1
	}
	if state.NextPhotoID < 1 {
		state.NextPhotoID = 1
	}
	if state.NextExportID < 1 {
		state.NextExportID = 1
	}
	if state.Accounts == nil {
		state.Accounts = []domain.Account{}
	}
	if state.Rooms == nil {
		state.Rooms = []domain.Room{}
	}
	if state.Exports == nil {
		state.Exports = []domain.ExportResult{}
	}
	return state
}
