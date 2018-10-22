// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	otp "github.com/hgfischer/go-otp"
)

type KeyType int8

const (
	// HOTP_TOKEN KeyType = 0
	TOTP_TOKEN KeyType = 1
)

type Key struct {
	Name     string       `json:"name,string"`
	Type     KeyType      `json:"type,int8"`
	Digits   int          `json:"digits,int"`
	Interval int          `json:"interval,int"`
	Counter  int          `json:"counter,int"`
	secret   SecretString `json:"-"`
}

func NewKey(name string) Key {
	return Key{
		Name:     name,
		Type:     TOTP_TOKEN,
		Digits:   6,
		Interval: 30,
		Counter:  1,
		secret:   newSecretString(name),
	}
}

func KeyFromStorage(storage Storage, name string) Key {
	value, err := storage.GetKey(name)
	if err != nil {
		fmt.Errorf("%s", err)
	}
	key := Key{}
	err = json.Unmarshal([]byte(value), &key)
	if err != nil {
		fmt.Errorf("%s", err)
	}
	key.secret = newSecretString(name)
	return key
}

func (k Key) String() string {
	return fmt.Sprintf("%s", k.Name)
}

func (k Key) VerboseString() string {
	return fmt.Sprintf("%s \t %d digits every %d seconds", k.Name, k.Digits, k.Interval)
}

func (k *Key) GenerateToken() string {
	switch k.Type {
	case TOTP_TOKEN:
		return k.totpToken()
	// case HOTP_TOKEN:
	// 	return k.hotpToken()
	default:
		panic("Unknown key type. Valid type: TOTP")
		// panic("Unknown key type. Valid type: TOTP or HOTP")
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

// // Generate a new HOTP token and increament counter
// func (k *Key) hotpToken() int {
// 	currentCounter := k.Counter

// 	secret, err := k.secret.Value()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	byteKey, err := base32StringToByte(string(secret))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	token := totp.Generate(byteKey, k.Digits, currentCounter)
// 	k.Counter++
// 	return token
// }

func (k *Key) Secret(secret string) error {
	return k.secret.Set([]byte(secret))
}

func (k *Key) Delete() error {
	return k.secret.Remove()
}

// func (this Key) MarshalJSON() ([]byte, error) {
// 	m := map[string]interface{}{} // ideally use make with the right capacity
// 	m["digits"] = this.Digits
// 	m["interval"]
// return json.Marshal(map[string]interface{}{
//		"some_field": w.SomeField,
//	})
// 	return json.Marshal(m)
// }

// func (k *Key) UnmarshalJSON(data []byte) error {
// 	var rawStrings map[string]interface{}

// 	if err := json.Unmarshal(data, &rawStrings); err != nil {
// 		return err
// 	}

// 	k.Name = rawStrings["name"].(string)
// 	k.Digits = int(rawStrings["digits"].(float64))
// 	k.Interval = int(rawStrings["interval"].(float64))
// 	return nil
// }

// func newKeyFromJSON(value []byte) (Key, error) {
// 	key := Key{}
// 	err := json.Unmarshal([]byte(value), &key)
// 	if err != nil {
// 		return Key{}, errors.Wrap(err, "cannot unmarshal")
// 	}
// 	key.Secret = NewSecretString("two-factor-authenticator", key.Name)
// 	return key, nil
// }
