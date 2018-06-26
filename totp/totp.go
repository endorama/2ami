// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package totp

import (
	"time"

	"github.com/endorama/two-factor-authenticator/hotp"
)

// Generate a new TOTP token based on current time
func Generate(key []byte, digits int, interval int) int {
	return totp(key, time.Now(), digits, interval)
}

// totp generate a TOTP token. Beware that length could be less than digits
func totp(key []byte, t time.Time, digits int, interval int) int {
	timeInterval := uint64(time.Duration(interval) * time.Second)
	counter := uint64(t.UnixNano()) / timeInterval
	return hotp.Generate(key, digits, counter)
}
