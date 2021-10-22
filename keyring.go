package main

import (
	"fmt"

	"github.com/99designs/keyring"
)

func openKeyring() (keyring.Keyring, error) {
	config := keyring.Config{
		AllowedBackends: []keyring.BackendType{
			keyring.SecretServiceBackend,
			keyring.KeychainBackend,
			keyring.WinCredBackend,
		},
		ServiceName:             "2ami",
		LibSecretCollectionName: "login",
	}
	ring, err := keyring.Open(config)
	if err != nil {
		return nil, fmt.Errorf("cannot open keyring: %w", err)
	}
	return ring, nil
}
