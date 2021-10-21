// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/99designs/keyring"
	otp "github.com/hgfischer/go-otp"
)

// func TestUnmarshalJSON(t *testing.T) {
// 	jsonString := "{\"name\": \"Test\", \"digits\": 8, \"interval\": 30}"

// 	data := Key{}
// 	err := json.Unmarshal([]byte(jsonString), &data)
// 	if err != nil {
// 		fmt.Errorf("%s", err)
// 	}

// 	test := Key{
// 		Name:     "Test",
// 		Digits:   8,
// 		Interval: 30,
// 	}

// 	if !reflect.DeepEqual(data, test) {
// 		t.Errorf("struct is not as expected: %+v %+v", data, test)
// 	}
// }

func _generateTotp(secret string, digits int) string {
	now := time.Now()
	interval := 30
	timeInterval := uint64(time.Duration(interval) * time.Second)
	counter := uint64(now.UnixNano()) / timeInterval

	token := &otp.HOTP{
		Secret:         string(secret),
		Length:         uint8(digits),
		Counter:        counter,
		IsBase32Secret: true,
	}
	return token.Get()
}

// openTestKeyring replaces openKeyring function to use a file backend instead
// This should allow running tests on all platforms without any OS specific
// dependency.
func openTestKeyring() (keyring.Keyring, error){
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	config := keyring.Config{
		AllowedBackends: []keyring.BackendType{
			keyring.FileBackend,
		},
		ServiceName:             "2ami",
		FilePasswordFunc: func(prompt string) (string, error) { 
			return "password", nil
		},
		FileDir: path,
	}
	ring, err := keyring.Open(config)
	if err != nil {
		return nil, fmt.Errorf("can't open keyring: %w", err)
	}
	return ring, nil
}

func TestKeyGenerateTotp(t *testing.T) {
	var generatedToken string

	ring, _ := openTestKeyring()

	key := NewKey(ring, "test")
	secretValue := "ORSXG5A="
	_ = key.Secret(secretValue)
	token := _generateTotp(secretValue, key.Digits)
	generatedToken = key.GenerateToken()
	if generatedToken != token {
		t.Errorf("Wrong token. Expected %s Actual %s", token, generatedToken)
	}

	key = NewKey(ring, "anothertest")
	secretValue = "MFXG65DIMVZHIZLTOQFA===="
	_ = key.Secret(secretValue)
	token = _generateTotp(secretValue, key.Digits)
	generatedToken = key.GenerateToken()
	if generatedToken != token {
		t.Errorf("Wrong token. Expected %s Actual %s", token, generatedToken)
	}

	key = NewKey(ring, "thisisatest2")
	secretValue = "ORUGS43JON2GK43UGI======"
	_ = key.Secret(secretValue)
	token = _generateTotp(secretValue, key.Digits)
	generatedToken = key.GenerateToken()
	if generatedToken != token {
		t.Errorf("Wrong token. Expected %s Actual %s", token, generatedToken)
	}
}
