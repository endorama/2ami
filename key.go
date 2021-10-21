// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/99designs/keyring"
	otp "github.com/hgfischer/go-otp"
)

type KeyType int8

const (
	HOTP_TOKEN KeyType = 0
	TOTP_TOKEN KeyType = 1
)

type Key struct {
	Name     string       `json:"name"`
	Type     KeyType      `json:"type,int8"` //nolint
	Digits   int          `json:"digits,int"` //nolint
	Interval int          `json:"interval,int"` //nolint
	Counter  int          `json:"counter,int"` //nolint
	secret   SecretString 
}

func NewKey(ring keyring.Keyring, name string) Key {
	return Key{
		Name:     name,
		Type:     TOTP_TOKEN,
		Digits:   6,
		Interval: 30,
		Counter:  1,
		secret:   newSecretString(name, ring),
	}
}

func KeyFromStorage(storage Storage, ring keyring.Keyring, name string) Key {
	value, err := storage.GetKey(name)
	if err != nil {
		fmt.Println(fmt.Errorf("%s", err))
	}
	key := Key{}
	err = json.Unmarshal([]byte(value), &key)
	if err != nil {
		fmt.Println(fmt.Errorf("%s", err))
	}
	key.secret = newSecretString(name, ring)
	return key
}

func (k Key) String() string {
	return k.Name
}

func (k Key) VerboseString() string {
	return fmt.Sprintf("%s \t %d digits every %d seconds", k.Name, k.Digits, k.Interval)
}

func (k *Key) GenerateToken() string {
	switch k.Type {
	case TOTP_TOKEN:
		return k.totpToken()
	case HOTP_TOKEN:
		return k.hotpToken()
	default:
		panic("Unknown key type. Valid type: TOTP or HOTP")
	}
}

func (k *Key) ExpiresIn() int {
	currentTime := time.Now()
	return k.Interval - (currentTime.Second() % k.Interval)
}

func (k *Key) totpToken() string {
	secret, err := k.secret.Value()
	if err != nil {
		log.Fatal(err)
	}
	totp := &otp.TOTP{
		Secret:         string(secret),
		Length:         uint8(k.Digits),
		Period:         uint8(k.Interval),
		IsBase32Secret: true,
	}
	token := totp.Get()
	return token
}

// Generate a new HOTP token and increament counter
func (k *Key) hotpToken() string {
	secret, err := k.secret.Value()
	if err != nil {
		log.Fatal(err)
	}
	hotp := &otp.HOTP{
		Secret:         string(secret),
		Counter:        uint64(k.Counter),
		Length:         uint8(k.Digits),
		IsBase32Secret: true,
	}
	token := hotp.Get()
	k.Counter++
	return token
}

func (k *Key) Secret(secret string) error {
	return k.secret.Set([]byte(secret))
}

func (k *Key) Delete() error {
	return k.secret.Remove()
}

func (k *Key) Rename(newName string) error {
	err := k.secret.Rename(newName)
	if err != nil {
		return err
	}
	k.Name = newName
	return nil
}
