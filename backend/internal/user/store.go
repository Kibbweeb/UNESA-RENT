package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	ErrNotFound   = errors.New("user tidak ditemukan")
	ErrEmailTaken = errors.New("email sudah terdaftar")
)

type Store interface {
	Create(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
}

type MemoryStore struct {
	mu          sync.RWMutex
	byID        map[string]User
	byEmail     map[string]User
	persistPath string
}

func NewMemoryStore(persistPath string) (*MemoryStore, error) {
	s := &MemoryStore{
		byID:        make(map[string]User),
		byEmail:     make(map[string]User),
		persistPath: strings.TrimSpace(persistPath),
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *MemoryStore) Create(_ context.Context, u User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	emailKey := normalizeEmail(u.Email)
	if _, exists := s.byEmail[emailKey]; exists {
		return ErrEmailTaken
	}

	s.byID[u.ID] = u
	s.byEmail[emailKey] = u

	if err := s.persistLocked(); err != nil {
		delete(s.byID, u.ID)
		delete(s.byEmail, emailKey)
		return fmt.Errorf("gagal menyimpan user: %w", err)
	}
	return nil
}

func (s *MemoryStore) FindByEmail(_ context.Context, email string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.byEmail[normalizeEmail(email)]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func (s *MemoryStore) FindByID(_ context.Context, id string) (User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.byID[id]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

func (s *MemoryStore) load() error {
	if s.persistPath == "" {
		return nil
	}

	raw, err := os.ReadFile(s.persistPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("membaca file user: %w", err)
	}
	if len(raw) == 0 {
		return nil
	}

	var users []User
	if err := json.Unmarshal(raw, &users); err != nil {
		return fmt.Errorf("file user tidak valid: %w", err)
	}
	for _, u := range users {
		s.byID[u.ID] = u
		s.byEmail[normalizeEmail(u.Email)] = u
	}
	return nil
}

func (s *MemoryStore) persistLocked() error {
	if s.persistPath == "" {
		return nil
	}

	users := make([]User, 0, len(s.byID))
	for _, u := range s.byID {
		users = append(users, u)
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.persistPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".users-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	defer os.Remove(tmpName)

	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.persistPath)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NewID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("gagal membuat id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
