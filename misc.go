// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"log"
	"os/user"
	"strconv"

	"github.com/pkg/errors"
)

func getUserHomeFolder() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", errors.Errorf("cannot get home folder: %q", err)
	}
	return usr.HomeDir, nil
}

func convertStringToInt(value string) (returnValue int) {
	returnValue, err := strconv.Atoi(value)
	if err != nil {
		log.Fatal(err)
	}
	return returnValue
}
