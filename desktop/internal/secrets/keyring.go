package secrets

import (
	"errors"
	"sync"

	"github.com/zalando/go-keyring"
)

const serviceName = "CompostIOT-Desktop"

var ErrNotFound = errors.New("token no encontrado en el llavero")

// TokenStore abstrae el almacenamiento seguro de tokens Bearer por servidor.
type TokenStore interface {
	SetToken(serverID, token string) error
	GetToken(serverID string) (string, error)
	DeleteToken(serverID string) error
}

// KeyringStore usa el llavero nativo del sistema operativo:
// Windows Credential Manager, macOS Keychain o Linux Secret Service.
type KeyringStore struct{}

func NewKeyringStore() *KeyringStore {
	return &KeyringStore{}
}

func (k *KeyringStore) SetToken(serverID, token string) error {
	return keyring.Set(serviceName, "token:"+serverID, token)
}

func (k *KeyringStore) GetToken(serverID string) (string, error) {
	val, err := keyring.Get(serviceName, "token:"+serverID)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	return val, err
}

func (k *KeyringStore) DeleteToken(serverID string) error {
	err := keyring.Delete(serviceName, "token:"+serverID)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

// MemoryStore almacena tokens en memoria (útil para pruebas automatizadas).
type MemoryStore struct {
	mu     sync.RWMutex
	tokens map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tokens: make(map[string]string)}
}

func (m *MemoryStore) SetToken(serverID, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[serverID] = token
	return nil
}

func (m *MemoryStore) GetToken(serverID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.tokens[serverID]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (m *MemoryStore) DeleteToken(serverID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, serverID)
	return nil
}
