// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"testing"
	"time"

	"github.com/endorama/two-factor-authenticator/hotp"
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

func _generateTotp(secret string, digits int) int {
	byteSecret, _ := base32StringToByte(secret)

	now := time.Now()
	interval := 30
	timeInterval := uint64(time.Duration(interval) * time.Second)
	counter := uint64(now.UnixNano()) / timeInterval

	return hotp.Generate(byteSecret, digits, counter)
}

func TestKeyGenerateTotp(t *testing.T) {
	key := NewKey("test")
	secretValue := "ORSXG5A="
	key.Secret(secretValue)
	token := _generateTotp(secretValue, key.Digits)
	generatedToken := key.GenerateToken()
	if generatedToken != token {
		t.Errorf("Wrong token. Expected %d Actual %d", token, generatedToken)
	}

	key = NewKey("anothertest")
	secretValue = "MFXG65DIMVZHIZLTOQFA===="
	key.Secret(secretValue)
	token = _generateTotp(secretValue, key.Digits)
	generatedToken = key.GenerateToken()
	if generatedToken != token {
		t.Errorf("Wrong token. Expected %d Actual %d", token, generatedToken)
	}

	key = NewKey("thisisatest2")
	secretValue = "ORUGS43JON2GK43UGI======"
	key.Secret(secretValue)
	token = _generateTotp(secretValue, key.Digits)
	generatedToken = key.GenerateToken()
	if generatedToken != token {
		t.Errorf("Wrong token. Expected %d Actual %d", token, generatedToken)
	}
}
