// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"encoding/base32"
	"log"
	"strconv"
	"strings"
)

func convertStringToInt(value string) (returnValue int, err error) {
	returnValue, err = strconv.Atoi(value)
	if err != nil {
		return returnValue, err
	}

	return returnValue, nil
}

// nolint
func base32StringToByte(data string) ([]byte, error) {
	return base32.StdEncoding.DecodeString(strings.ToUpper(data))
}

func debugPrint(v string) {
	if debug {
		log.Println(v)
	}
}
