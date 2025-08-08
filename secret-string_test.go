// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"bytes"
	"testing"
)

func TestSecretString_Set(t *testing.T) {
	ring, err := openTestKeyring(t)
	if err != nil {
		t.Fatalf("Failed to open test keyring: %v", err)
	}

	tests := []struct {
		name    string
		keyName string
		data    []byte
		wantErr bool
	}{
		{
			name:    "set valid data",
			keyName: "test-key-1",
			data:    []byte("test-secret-data"),
			wantErr: false,
		},
		{
			name:    "set empty data",
			keyName: "test-key-2",
			data:    []byte(""),
			wantErr: false,
		},
		{
			name:    "set binary data",
			keyName: "test-key-3",
			data:    []byte{0x00, 0x01, 0x02, 0xFF},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSecretString(tt.keyName, ring)
			t.Cleanup(func() { _ = s.Remove() })

			err := s.Set(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecretString.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecretString_Value(t *testing.T) {
	ring, err := openTestKeyring(t)
	if err != nil {
		t.Fatalf("Failed to open test keyring: %v", err)
	}

	tests := []struct {
		name      string
		keyName   string
		setupData []byte
		setup     bool
		wantData  []byte
		wantErr   bool
	}{
		{
			name:      "get existing data",
			keyName:   "test-key-value-1",
			setupData: []byte("test-secret-data"),
			setup:     true,
			wantData:  []byte("test-secret-data"),
			wantErr:   false,
		},
		{
			name:      "get empty data",
			keyName:   "test-key-value-2",
			setupData: []byte(""),
			setup:     true,
			wantData:  []byte(""),
			wantErr:   false,
		},
		{
			name:      "get binary data",
			keyName:   "test-key-value-3",
			setupData: []byte{0x00, 0x01, 0x02, 0xFF},
			setup:     true,
			wantData:  []byte{0x00, 0x01, 0x02, 0xFF},
			wantErr:   false,
		},
		{
			name:    "get non-existent key",
			keyName: "non-existent-key",
			setup:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSecretString(tt.keyName, ring)

			if tt.setup {
				err := s.Set(tt.setupData)
				if err != nil {
					t.Fatalf("Failed to setup test data: %v", err)
				}
				t.Cleanup(func() { _ = s.Remove() })
			}

			got, err := s.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("SecretString.Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !bytes.Equal(got, tt.wantData) {
				t.Errorf("SecretString.Value() = %v, want %v", got, tt.wantData)
			}
		})
	}
}

func TestSecretString_Remove(t *testing.T) {
	ring, err := openTestKeyring(t)
	if err != nil {
		t.Fatalf("Failed to open test keyring: %v", err)
	}

	tests := []struct {
		name      string
		keyName   string
		setupData []byte
		setup     bool
		wantErr   bool
	}{
		{
			name:      "remove existing key",
			keyName:   "test-key-remove-1",
			setupData: []byte("test-data"),
			setup:     true,
			wantErr:   false,
		},
		{
			name:    "remove non-existent key",
			keyName: "non-existent-remove-key",
			setup:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSecretString(tt.keyName, ring)

			if tt.setup {
				err := s.Set(tt.setupData)
				if err != nil {
					t.Fatalf("Failed to setup test data: %v", err)
				}
			}

			err := s.Remove()
			if (err != nil) != tt.wantErr {
				t.Errorf("SecretString.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify the key is actually removed
			if !tt.wantErr {
				_, err = s.Value()
				if err == nil {
					t.Error("Expected key to be removed, but it still exists")
				}
			}
		})
	}
}

func TestSecretString_Rename(t *testing.T) {
	ring, err := openTestKeyring(t)
	if err != nil {
		t.Fatalf("Failed to open test keyring: %v", err)
	}

	tests := []struct {
		name      string
		keyName   string
		newName   string
		setupData []byte
		setup     bool
		wantErr   bool
	}{
		{
			name:      "rename existing key",
			keyName:   "test-key-rename-1",
			newName:   "test-key-renamed-1",
			setupData: []byte("test-data-for-rename"),
			setup:     true,
			wantErr:   false,
		},
		{
			name:    "rename non-existent key",
			keyName: "non-existent-rename-key",
			newName: "new-name",
			setup:   false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSecretString(tt.keyName, ring)

			if tt.setup {
				err := s.Set(tt.setupData)
				if err != nil {
					t.Fatalf("Failed to setup test data: %v", err)
				}
			}

			err := s.Rename(tt.newName)
			if (err != nil) != tt.wantErr {
				t.Errorf("SecretString.Rename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the data exists with the new name
				if s.Name != tt.newName {
					t.Errorf("SecretString.Name = %v, want %v", s.Name, tt.newName)
				}

				got, err := s.Value()
				if err != nil {
					t.Errorf("Failed to get data after rename: %v", err)
				} else if !bytes.Equal(got, tt.setupData) {
					t.Errorf("Data after rename = %v, want %v", got, tt.setupData)
				}

				// Verify the old key is removed
				oldS := newSecretString(tt.keyName, ring)
				_, err = oldS.Value()
				if err == nil {
					t.Error("Expected old key to be removed, but it still exists")
				}

				// Clean up
				t.Cleanup(func() { _ = s.Remove() })
			}
		})
	}
}

func TestNewSecretString(t *testing.T) {
	ring, err := openTestKeyring(t)
	if err != nil {
		t.Fatalf("Failed to open test keyring: %v", err)
	}

	name := "test-secret"
	s := newSecretString(name, ring)

	if s.Name != name {
		t.Errorf("newSecretString() Name = %v, want %v", s.Name, name)
	}
	if s.ring != ring {
		t.Error("newSecretString() ring not set correctly")
	}
}

func TestSecretString_SetAndValueRoundTrip(t *testing.T) {
	ring, err := openTestKeyring(t)
	if err != nil {
		t.Fatalf("Failed to open test keyring: %v", err)
	}

	testData := []byte("round-trip-test-data")
	keyName := "round-trip-test"

	s := newSecretString(keyName, ring)
	t.Cleanup(func() { _ = s.Remove() })

	// Set data
	err = s.Set(testData)
	if err != nil {
		t.Fatalf("Failed to set data: %v", err)
	}

	// Get data back
	got, err := s.Value()
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if !bytes.Equal(got, testData) {
		t.Errorf("Round trip failed: got %v, want %v", got, testData)
	}
}
