// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"fmt"
	"log"

	"github.com/99designs/keyring"
	"github.com/pkg/errors"
)

type SecretString struct {
	Name string
	ring keyring.Keyring
}

func newSecretString(name string) SecretString {
	ring, err := keyring.Open(keyring.Config{
		AllowedBackends:         []keyring.BackendType{keyring.SecretServiceBackend},
		ServiceName:             "two-factor-authenticator",
		LibSecretCollectionName: "login",
	})
	if err != nil {
		log.Fatalf("can't open keyring: %s", err)
	}
	return SecretString{
		Name: name,
		ring: ring,
	}
}

func (s *SecretString) Set(data []byte) error {
	item := keyring.Item{
		Key:         s.Name,
		Label:       s.Name,
		Data:        data,
		Description: fmt.Sprintf("2FA key for %s", s.Name),
	}
	err := s.ring.Set(item)
	if err != nil {
		return errors.Wrap(err, "cannot set data from keyring")
	}
	return nil
}

func (s *SecretString) Value() ([]byte, error) {
	fmt.Println(s.Name)
	i, err := s.ring.Get(s.Name)
	if err != nil {
		return []byte{}, errors.Wrap(err, "cannot get data from keyring")
	}
	return i.Data, nil
}
