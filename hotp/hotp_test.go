// Copyright 2018 Edoardo Tenani. All rights reserved.
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package hotp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"log"
	"strings"
	"testing"
	"time"
)

func _hotp(key []byte, counter uint64, digits int) int {
	h := hmac.New(sha1.New, key)
	binary.Write(h, binary.BigEndian, counter)
	sum := h.Sum(nil)
	v := binary.BigEndian.Uint32(sum[sum[len(sum)-1]&0x0F:]) & 0x7FFFFFFF
	d := uint32(1)
	for i := 0; i < digits && i < 8; i++ {
		d *= 10
	}
	return int(v % d)
}

func toByte(key string) []byte {
	decodedKey, err := base32.StdEncoding.DecodeString(strings.ToUpper(key))
	if err != nil {
		log.Fatal(err)
	}
	return decodedKey
}

func TestHotp(t *testing.T) {
	// now in the past, to get known tokens
	location, _ := time.LoadLocation("Europe/Rome")
	now := time.Date(2018, 7, 4, 22, 30, 15, 215234965, location)
	interval := 30
	timeInterval := uint64(time.Duration(interval) * time.Second)
	counter := uint64(now.UnixNano()) / timeInterval

	token := Generate(toByte("ORSXG5A="), 6, counter)
	if token != 193637 {
		t.Errorf("Invalid token generated. Expected %d Actual %d", 193637, token)
	}

	token = Generate(toByte("MFXG65DIMVZHIZLTOQFA===="), 6, counter)
	if token != 463293 {
		t.Errorf("Invalid token generated. Expected %d Actual %d", 463293, token)
	}

	token = Generate(toByte("ORUGS43JON2GK43UGI======"), 6, counter)
	if token != 335172 {
		t.Errorf("Invalid token generated. Expected %d Actual %d", 335172, token)
	}
}
