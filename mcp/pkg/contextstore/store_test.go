package contextstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestContextStore_SaveLoad(t *testing.T) {
	// Create a temporary file for the store
	tmpStoreFile := filepath.Join(os.TempDir(), "test_context_store.json")
	defer os.Remove(tmpStoreFile)

	// Create a temporary file to be added to context
	tmpContentFile, err := os.CreateTemp("", "test_content_*.go")
	if err != nil {
		t.Fatalf("Failed to create temp content file: %v", err)
	}
	defer os.Remove(tmpContentFile.Name())
	if _, err := tmpContentFile.WriteString("package main\nfunc main() {}"); err != nil {
		t.Fatal(err)
	}
	tmpContentFile.Close()
	absPath, _ := filepath.Abs(tmpContentFile.Name())

	// Initialize store
	store := NewContextStore(tmpStoreFile)
	sessionID := "test-session"

	// Create a session and add an item
	sess := store.GetSession(sessionID)
	err = sess.AddFile(absPath)
	if err != nil {
		t.Fatalf("Failed to add file: %v", err)
	}

	// Verify file is in memory
	items := sess.ListItems()
	if len(items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(items))
	}

	// Verify persistence
	// AddFile calls `go s.store.Save()`, so we need to ensure it saves.
	// Since we can't reliably wait for the async goroutine without hooks,
	// we call Save() manually for the purpose of this persistence test.
	err = store.Save()
	if err != nil {
		t.Fatalf("Failed to save store: %v", err)
	}

	// Create a new store instance from the same file
	newStore := NewContextStore(tmpStoreFile)

	// Verify data was loaded
	newSess := newStore.GetSession(sessionID)
	loadedItems := newSess.ListItems()
	if len(loadedItems) != 1 {
		t.Errorf("Expected 1 persisted item, got %d", len(loadedItems))
	}
	if loadedItems[0].Path != absPath {
		t.Errorf("Expected path %s, got %s", absPath, loadedItems[0].Path)
	}
}

func TestContextStore_Concurrency(t *testing.T) {
	tmpStoreFile := filepath.Join(os.TempDir(), "test_context_store_concurrent.json")
	defer os.Remove(tmpStoreFile)

	// Create a dummy file to read
	tmpContentFile, _ := os.CreateTemp("", "dummy_*.txt")
	tmpContentFile.WriteString("dummy content")
	tmpContentFile.Close()
	defer os.Remove(tmpContentFile.Name())
	absPath, _ := filepath.Abs(tmpContentFile.Name())

	store := NewContextStore(tmpStoreFile)
	sessionID := "concurrent-session"
	sess := store.GetSession(sessionID)

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent Adds (using the same file, but that's fine for race detection on map)
	// Actually, let's look at AddFile implementation:
	// sc.Items[absPath] = ...
	// If we use the same path, we are overwriting the same key.
	// This is still a valid race test for the map write.
	// To test map resizing/collisions, strictly we'd want different keys, but `AddFile` takes a real path.
	// We can't easily generate 50 real files in a test efficiently, so we'll reuse the file.
	// The race detector will still catch if `sc.Items` is accessed unsafely.

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = sess.AddFile(absPath)
		}(i)
	}

	// Concurrent Saves
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.Save()
		}()
	}

	wg.Wait()

	items := sess.ListItems()
	if len(items) == 0 {
		t.Errorf("Expected items, got 0")
	}
}

func TestContextStore_JSONMarshaling(t *testing.T) {
	// regression test for race condition in MarshalJSON
	store := NewContextStore("")
	sessionID := "marshal-test"
	sess := store.GetSession(sessionID)

	// Manually inject an item to avoid disk I/O
	sess.Items["file1"] = ContextItem{Path: "file1", TokenSize: 10}

	// Safely marshal
	data, err := json.Marshal(sess)
	if err != nil {
		t.Fatalf("Failed to marshal session: %v", err)
	}

	// verify content
	var unmarshaled SessionContext
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(unmarshaled.Items) != 1 {
		t.Errorf("Expected 1 item after unmarshal, got %d", len(unmarshaled.Items))
	}
}
