// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import "testing"

func Test(t *testing.T) {}

func TestIsValidBase32_whitValidData(t *testing.T) {
	if isValidBase32("ORSXG5A=") != nil {
		t.Error("Not a valid Base32 string")
	}
}

func TestIsValidBase32_whitInvalidData(t *testing.T) {
	if isValidBase32("*^ASD") == nil {
		t.Error("base32 validator missed this invalid base32 string")
	}
}
