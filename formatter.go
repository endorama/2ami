// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"fmt"
	"strconv"
)

func tokenFormatter(formatter string, digits, token int) string {
	var output string
	switch formatter {
	case "google-authenticator":
		output = formatter_googleAuthenticator(digits, token)
	case "default":
		output = strconv.Itoa(token)
	}
	return output
}

func formatter_googleAuthenticator(digits, token int) string {
	output := strconv.Itoa(token)
	layout := fmt.Sprintf("%%0%dd", digits)
	if len(output) < digits {
		return fmt.Sprintf(layout, token)
	}
	return output
}
