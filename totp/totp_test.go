// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package totp

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// _generateToken creates a token with predefined key
func _generateToken(myTime time.Time, digits int, interval int) int {
	return totp([]byte("test"), myTime, digits, interval)
}

// _generateFutureTokens can be used to generate tokens in the future.
func _generateFutureTokens() {
	myTime := time.Now()
	for i := 0; i < 5; i++ {
		fmt.Printf("%s %d\n", myTime, _generateToken(myTime, 6, 30))
		myTime = myTime.Add(30 * time.Second)
	}
	myTime = time.Now()
	for i := 0; i < 5; i++ {
		fmt.Printf("%s %d\n", myTime, _generateToken(myTime, 6, 60))
		myTime = myTime.Add(60 * time.Second)
	}
	myTime = time.Now()
	for i := 0; i < 5; i++ {
		fmt.Printf("%s %d\n", myTime, _generateToken(myTime, 10, 30))
		myTime = myTime.Add(30 * time.Second)
	}
	os.Exit(1)
}
func TestTotp(t *testing.T) {

	// _generateFutureTokens() // uncomment to generate, will Exit 1

	type TestCase struct {
		Digits        int
		Interval      int
		Time          string
		ExpectedToken int
	}

	var testCases = []TestCase{
		TestCase{Digits: 6, Interval: 30, Time: "2018-06-27 00:15:09.075934839 +0200", ExpectedToken: 955788},
		TestCase{Digits: 6, Interval: 30, Time: "2018-06-27 00:15:39.075934839 +0200", ExpectedToken: 985695},
		TestCase{Digits: 6, Interval: 30, Time: "2018-06-27 00:16:09.075934839 +0200", ExpectedToken: 200922},
		TestCase{Digits: 6, Interval: 30, Time: "2018-06-27 00:16:39.075934839 +0200", ExpectedToken: 972657},
		TestCase{Digits: 6, Interval: 30, Time: "2018-06-27 00:17:09.075934839 +0200", ExpectedToken: 236324},

		TestCase{Digits: 6, Interval: 60, Time: "2018-06-27 00:15:09.076282471 +0200", ExpectedToken: 386889},
		TestCase{Digits: 6, Interval: 60, Time: "2018-06-27 00:16:09.076282471 +0200", ExpectedToken: 312504},
		TestCase{Digits: 6, Interval: 60, Time: "2018-06-27 00:17:09.076282471 +0200", ExpectedToken: 20257},
		TestCase{Digits: 6, Interval: 60, Time: "2018-06-27 00:18:09.076282471 +0200", ExpectedToken: 198545},
		TestCase{Digits: 6, Interval: 60, Time: "2018-06-27 00:19:09.076282471 +0200", ExpectedToken: 105702},

		TestCase{Digits: 10, Interval: 30, Time: "2018-06-27 00:15:09.076444814 +0200", ExpectedToken: 57955788},
		TestCase{Digits: 10, Interval: 30, Time: "2018-06-27 00:15:39.076444814 +0200", ExpectedToken: 14985695},
		TestCase{Digits: 10, Interval: 30, Time: "2018-06-27 00:16:09.076444814 +0200", ExpectedToken: 8200922},
		TestCase{Digits: 10, Interval: 30, Time: "2018-06-27 00:16:39.076444814 +0200", ExpectedToken: 44972657},
		TestCase{Digits: 10, Interval: 30, Time: "2018-06-27 00:17:09.076444814 +0200", ExpectedToken: 28236324},
	}

	timeLayout := "2006-01-02 15:04:05.000000000 -0700"

	for _, v := range testCases {
		thisTime, _ := time.Parse(timeLayout, v.Time)
		token := _generateToken(thisTime, v.Digits, v.Interval)
		if token != v.ExpectedToken {
			t.Errorf("token differs. Expected <%d> Actual <%d>", token, v.ExpectedToken)
		}
	}
}
