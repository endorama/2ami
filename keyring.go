package main

import (
	"fmt"

	"github.com/99designs/keyring"
	"github.com/spf13/viper"
)

func openKeyring() (keyring.Keyring, error) {
	config := keyring.Config{
		AllowedBackends: []keyring.BackendType{
			keyring.SecretServiceBackend,
			keyring.KeychainBackend,
			keyring.WinCredBackend,
		},
		ServiceName:             "2ami",
		KeychainName:            viper.GetString("ring"),
		LibSecretCollectionName: viper.GetString("ring"),
		WinCredPrefix:           viper.GetString("ring"),
	}
	ring, err := keyring.Open(config)
	if err != nil {
		return nil, fmt.Errorf("cannot open keyring: %w", err)
	}
	return ring, nil
}
