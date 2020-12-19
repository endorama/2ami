// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"testing"
)

func TestTokenFormatter(t *testing.T) {
	if formatted := tokenFormatter("default", 6, 123); formatted != "123" {
		t.Errorf("Wrong format for default formatter. Expected %s Actual %s", "123", formatted)
	}

	if formatted := tokenFormatter("google-authenticator", 6, 123); formatted != "000123" {
		t.Errorf("Wrong format for google-authenticator formatter. Expected %s Actual %s", "000123", formatted)
	}
}

func TestFormatter_googleAuthenticator(t *testing.T) {
	type TestCase struct {
		Digits   int
		Token    int
		Expected string
	}

	testCases := []TestCase{
		{Digits: 6, Token: 1, Expected: "000001"},
		{Digits: 6, Token: 12, Expected: "000012"},
		{Digits: 6, Token: 123, Expected: "000123"},
		{Digits: 6, Token: 1234, Expected: "001234"},
		{Digits: 6, Token: 12345, Expected: "012345"},
		{Digits: 6, Token: 123456, Expected: "123456"},
		{Digits: 8, Token: 1234567, Expected: "01234567"},
		{Digits: 8, Token: 12345678, Expected: "12345678"},
	}

	for _, v := range testCases {
		if formatted := formatter_googleAuthenticator(v.Digits, v.Token); formatted != v.Expected {
			t.Errorf("Wrong format. Expected %s Actual %s", v.Expected, formatted)
		}
	}
}
