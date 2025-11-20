// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

// setupTestStorage creates and initializes a storage instance for testing
func setupTestStorage(t *testing.T) (Storage, func()) {
	tmpDir := t.TempDir()
	filename := "test.db"

	storage := NewStorage(tmpDir, filename)
	err := storage.Init()
	if err != nil {
		t.Fatalf("Failed to initialize test storage: %v", err)
	}

	cleanup := func() {
		require.NoError(t, storage.Close())
	}

	return storage, cleanup
}

// setupTestStorageWithoutInit creates a storage instance without initializing it
func setupTestStorageWithoutInit(t *testing.T) Storage {
	tmpDir := t.TempDir()
	filename := "test.db"
	return NewStorage(tmpDir, filename)
}

func TestNewStorage(t *testing.T) {
	storage := setupTestStorageWithoutInit(t)

	if storage.folder == "" {
		t.Error("Expected folder to be set")
	}
	if storage.filename != "test.db" {
		t.Errorf("Expected filename test.db, got %s", storage.filename)
	}
	if storage.db != nil {
		t.Error("Expected db to be nil before Init()")
	}
}

func TestStorage_Init(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	if storage.db == nil {
		t.Error("Expected db to be initialized after Init()")
	}

	// Verify file was created
	dbPath := filepath.Join(storage.folder, storage.filename)
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file %s was not created", dbPath)
	}
}

func TestStorage_Init_InvalidPath(t *testing.T) {
	// Use a path that doesn't exist and can't be created
	invalidPath := "/nonexistent/folder"
	filename := "test.db"

	storage := NewStorage(invalidPath, filename)

	err := storage.Init()
	if err == nil {
		require.NoError(t, storage.Close())
		t.Error("Expected Init() to fail with invalid path")
	}
}

func TestStorage_Close(t *testing.T) {
	storage, _ := setupTestStorage(t)

	// Should not panic
	require.NoError(t, storage.Close())

	// Multiple closes should not panic
	require.NoError(t, storage.Close())
}

func TestStorage_AddKey(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	key := "testkey"
	value := []byte("testvalue")

	success, err := storage.AddKey(key, value)
	if err != nil {
		t.Errorf("AddKey() failed: %v", err)
	}
	if !success {
		t.Error("AddKey() returned false for successful operation")
	}

	// Verify key was added by retrieving it
	retrievedValue, err := storage.GetKey(key)
	if err != nil {
		t.Errorf("GetKey() failed: %v", err)
	}
	if !reflect.DeepEqual(retrievedValue, value) {
		t.Errorf("Expected value %v, got %v", value, retrievedValue)
	}
}

func TestStorage_AddKey_OverwriteExisting(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	key := "testkey"
	value1 := []byte("testvalue1")
	value2 := []byte("testvalue2")

	// Add first value
	_, err := storage.AddKey(key, value1)
	if err != nil {
		t.Errorf("First AddKey() failed: %v", err)
	}

	// Overwrite with second value
	success, err := storage.AddKey(key, value2)
	if err != nil {
		t.Errorf("Second AddKey() failed: %v", err)
	}
	if !success {
		t.Error("AddKey() returned false for overwrite operation")
	}

	// Verify second value is stored
	retrievedValue, err := storage.GetKey(key)
	if err != nil {
		t.Errorf("GetKey() failed: %v", err)
	}
	if !reflect.DeepEqual(retrievedValue, value2) {
		t.Errorf("Expected value %v, got %v", value2, retrievedValue)
	}
}

func TestStorage_GetKey(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	key := "testkey"
	value := []byte("testvalue")

	// Add key first
	_, err := storage.AddKey(key, value)
	if err != nil {
		t.Errorf("AddKey() failed: %v", err)
	}

	// Get the key
	retrievedValue, err := storage.GetKey(key)
	if err != nil {
		t.Errorf("GetKey() failed: %v", err)
	}
	if !reflect.DeepEqual(retrievedValue, value) {
		t.Errorf("Expected value %v, got %v", value, retrievedValue)
	}
}

func TestStorage_GetKey_NonExistent(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Try to get a non-existent key
	value, err := storage.GetKey("nonexistent")
	if err != nil {
		t.Errorf("GetKey() failed for non-existent key: %v", err)
	}
	if value != nil {
		t.Errorf("Expected nil value for non-existent key, got %v", value)
	}
}

func TestStorage_ListKey_Empty(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	keys, err := storage.ListKey()
	if err != nil {
		t.Errorf("ListKey() failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected empty key list, got %v", keys)
	}
}

func TestStorage_ListKey_WithKeys(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	expectedKeys := []string{"key1", "key2", "key3"}

	// Add multiple keys
	for _, key := range expectedKeys {
		_, err := storage.AddKey(key, []byte(fmt.Sprintf("value_%s", key)))
		if err != nil {
			t.Errorf("AddKey() failed for key %s: %v", key, err)
		}
	}

	// List keys
	keys, err := storage.ListKey()
	if err != nil {
		t.Errorf("ListKey() failed: %v", err)
	}

	// Sort both slices for comparison since BoltDB doesn't guarantee order
	sort.Strings(keys)
	sort.Strings(expectedKeys)

	if !reflect.DeepEqual(keys, expectedKeys) {
		t.Errorf("Expected keys %v, got %v", expectedKeys, keys)
	}
}

func TestStorage_RemoveKey(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	key := "testkey"
	value := []byte("testvalue")

	// Add key first
	_, err := storage.AddKey(key, value)
	if err != nil {
		t.Errorf("AddKey() failed: %v", err)
	}

	// Remove the key
	err = storage.RemoveKey(key)
	if err != nil {
		t.Errorf("RemoveKey() failed: %v", err)
	}

	// Verify key is removed
	retrievedValue, err := storage.GetKey(key)
	if err != nil {
		t.Errorf("GetKey() failed after removal: %v", err)
	}
	if retrievedValue != nil {
		t.Errorf("Expected nil value after removal, got %v", retrievedValue)
	}

	// Verify key is not in list
	keys, err := storage.ListKey()
	if err != nil {
		t.Errorf("ListKey() failed: %v", err)
	}
	for _, k := range keys {
		if k == key {
			t.Errorf("Key %s still appears in list after removal", key)
		}
	}
}

func TestStorage_RemoveKey_NonExistent(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Try to remove a non-existent key (should not error)
	err := storage.RemoveKey("nonexistent")
	if err != nil {
		t.Errorf("RemoveKey() failed for non-existent key: %v", err)
	}
}

func TestStorage_Integration(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Test complete workflow
	testData := map[string][]byte{
		"user1": []byte("secret1"),
		"user2": []byte("secret2"),
		"user3": []byte("secret3"),
	}

	// Add all keys
	for key, value := range testData {
		success, err := storage.AddKey(key, value)
		if err != nil || !success {
			t.Errorf("Failed to add key %s: %v", key, err)
		}
	}

	// Verify all keys exist
	keys, err := storage.ListKey()
	if err != nil {
		t.Errorf("ListKey() failed: %v", err)
	}
	if len(keys) != len(testData) {
		t.Errorf("Expected %d keys, got %d", len(testData), len(keys))
	}

	// Verify all values
	for key, expectedValue := range testData {
		value, err := storage.GetKey(key)
		if err != nil {
			t.Errorf("GetKey() failed for %s: %v", key, err)
		}
		if !reflect.DeepEqual(value, expectedValue) {
			t.Errorf("Expected value %v for key %s, got %v", expectedValue, key, value)
		}
	}

	// Remove one key
	err = storage.RemoveKey("user2")
	if err != nil {
		t.Errorf("RemoveKey() failed: %v", err)
	}

	// Verify it's gone
	keys, err = storage.ListKey()
	if err != nil {
		t.Errorf("ListKey() failed: %v", err)
	}
	if len(keys) != len(testData)-1 {
		t.Errorf("Expected %d keys after removal, got %d", len(testData)-1, len(keys))
	}

	value, err := storage.GetKey("user2")
	if err != nil {
		t.Errorf("GetKey() failed for removed key: %v", err)
	}
	if value != nil {
		t.Errorf("Expected nil value for removed key, got %v", value)
	}
}

func TestStorage_EmptyValues(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Test with empty byte slice
	key := "emptykey"
	emptyValue := []byte{}

	success, err := storage.AddKey(key, emptyValue)
	if err != nil || !success {
		t.Errorf("AddKey() failed for empty value: %v", err)
	}

	retrievedValue, err := storage.GetKey(key)
	if err != nil {
		t.Errorf("GetKey() failed for empty value: %v", err)
	}
	if !reflect.DeepEqual(retrievedValue, emptyValue) {
		t.Errorf("Expected empty value, got %v", retrievedValue)
	}
}

func TestStorage_LargeValues(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Test with large value
	key := "largekey"
	largeValue := make([]byte, 10000)
	for i := range largeValue {
		largeValue[i] = byte(i % 256)
	}

	success, err := storage.AddKey(key, largeValue)
	if err != nil || !success {
		t.Errorf("AddKey() failed for large value: %v", err)
	}

	retrievedValue, err := storage.GetKey(key)
	if err != nil {
		t.Errorf("GetKey() failed for large value: %v", err)
	}
	if !reflect.DeepEqual(retrievedValue, largeValue) {
		t.Errorf("Large value was not stored/retrieved correctly")
	}
}
