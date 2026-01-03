package contextstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// ContextItem represents a file or content added to context
type ContextItem struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	TokenSize int    `json:"token_size"` // Estimated
}

// SessionContext holds the context for a specific session/user
type SessionContext struct {
	Items map[string]ContextItem `json:"items"`
	mu    sync.RWMutex
	store *ContextStore // Back reference for saving
}

// MarshalJSON implements custom marshaling to ensure thread safety
func (sc *SessionContext) MarshalJSON() ([]byte, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// Define a type alias to prevent recursion
	type Alias SessionContext
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(sc),
	})
}

// ContextStore manages context across multiple sessions
type ContextStore struct {
	Sessions map[string]*SessionContext `json:"sessions"`
	mu       sync.RWMutex
	filePath string
}

func NewContextStore(filePath string) *ContextStore {
	store := &ContextStore{
		Sessions: make(map[string]*SessionContext),
		filePath: filePath,
	}
	// Try loading existing state
	_ = store.Load()
	return store
}

// Load reads the context from disk
func (s *ContextStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Unmarshal into a temporary struct to avoid overwriting safe pointers or missing back-refs
	// Actually we can unmarshal directly into s.Sessions if we are careful
	// But JSON unmarshal wipes map? No, it unmarshals into it.
	// Simpler to unmarshal to a temp map
	var tempSessions map[string]*SessionContext
	if err := json.Unmarshal(data, &tempSessions); err != nil {
		return err
	}

	// Rehydrate sessions
	s.Sessions = tempSessions
	for _, sess := range s.Sessions {
		sess.store = s
		if sess.Items == nil {
			sess.Items = make(map[string]ContextItem)
		}
	}

	return nil
}

// Save writes the context to disk
func (s *ContextStore) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filePath == "" {
		return nil
	}

	data, err := json.MarshalIndent(s.Sessions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *ContextStore) GetSession(sessionID string) *SessionContext {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.Sessions[sessionID]; !ok {
		s.Sessions[sessionID] = &SessionContext{
			Items: make(map[string]ContextItem),
			store: s,
		}
	}
	return s.Sessions[sessionID]
}

// AddFile reads a file and adds it to the session context
func (sc *SessionContext) AddFile(path string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Check if file exists
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Simple token estimation (4 chars ~= 1 token)
	tokenSize := len(content) / 4

	sc.Items[absPath] = ContextItem{
		Path:      absPath,
		Content:   string(content),
		TokenSize: tokenSize,
	}

	// Trigger Save
	if sc.store != nil {
		go sc.store.Save() // Save asynchronously to avoid blocking
	}

	return nil
}

// RemoveFile removes a file from context
func (sc *SessionContext) RemoveFile(path string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err == nil {
		delete(sc.Items, absPath)
	}
	// Also try deleting as-is in case it was stored differently
	delete(sc.Items, path)

	// Trigger Save
	if sc.store != nil {
		go sc.store.Save()
	}
}

// ListItems returns all items in context
func (sc *SessionContext) ListItems() []ContextItem {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	items := make([]ContextItem, 0, len(sc.Items))
	for _, item := range sc.Items {
		items = append(items, item)
	}
	return items
}

// Clear removes all items
func (sc *SessionContext) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.Items = make(map[string]ContextItem)

	// Trigger Save
	if sc.store != nil {
		go sc.store.Save()
	}
}

// GetTotalTokens returns estimated total tokens
func (sc *SessionContext) GetTotalTokens() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	total := 0
	for _, item := range sc.Items {
		total += item.TokenSize
	}
	return total
}
