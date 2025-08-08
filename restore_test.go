package main

import (
	"testing"
)

func TestRestore_InvalidData(t *testing.T) {
	// Create a real storage instance for testing
	storage := NewStorage("/tmp", "test.db")

	// Test with invalid backup data
	err := restore(storage, "invalid_backup_data", "password", backupFormat2ami)
	if err == nil {
		t.Error("Expected error for invalid backup data, got nil")
	}
}

func TestRestore_EmptyData(t *testing.T) {
	// Create a real storage instance for testing
	storage := NewStorage("/tmp", "test.db")

	// Test with empty backup data
	err := restore(storage, "", "password", backupFormat2ami)
	if err == nil {
		t.Error("Expected error for empty backup data, got nil")
	}
}
