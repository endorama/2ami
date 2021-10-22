// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"fmt"

	"github.com/99designs/keyring"
	"github.com/pkg/errors"
)

// SecretString represent a named string in a secure storage
type SecretString struct {
	Name string
	ring keyring.Keyring
}

func newSecretString(name string, ring keyring.Keyring) SecretString {
	return SecretString{
		Name: name,
		ring: ring,
	}
}

// Set write value for the current string in the secure storage
// Can error if writing to secure storage fails
func (s *SecretString) Set(data []byte) error {
	item := keyring.Item{
		Key:         s.Name,
		Label:       s.Name,
		Data:        data,
		Description: fmt.Sprintf("2FA key for %s", s.Name),
	}
	err := s.ring.Set(item)
	if err != nil {
		return fmt.Errorf("cannot set data from keyring: %w", err)
	}
	return nil
}

// Value returns value of the current string, if present
// Can fail if reading from secure storage fails
func (s *SecretString) Value() ([]byte, error) {
	i, err := s.ring.Get(s.Name)
	if err != nil {
		return []byte{}, fmt.Errorf("cannot get data from keyring: %w", err)
	}
	if i.Data == nil {
		return []byte{}, errors.New("empty data from keyring; was the key removed from it?")
	}
	return i.Data, nil
}

// Remove delete current string from storage
func (s *SecretString) Remove() error {
	return s.ring.Remove(s.Name)
}

// Rename updates current secret string name
func (s *SecretString) Rename(name string) error {
	data, err := s.Value()
	if err != nil {
		return err
	}
	err = s.Remove()
	if err != nil {
		return err
	}
	s.Name = name
	err = s.Set(data)
	if err != nil {
		return err
	}
	return nil
}
