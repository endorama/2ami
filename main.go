// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"encoding/base32"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/OpenPeeDeeP/xdg"
	docopt "github.com/docopt/docopt.go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/atotto/clipboard"
	"github.com/mitchellh/cli"
)

var (
	debug   = false
	verbose = false
	version = "dev"
	ui      cli.Ui
)

func usage() string {
	return `Two factor authenticator for your command line.

Usage:
  2ami add <name> [--digits=<digits>] [--interval=<seconds>] [--verbose]
  2ami dump [<name>] [--verbose]
  2ami generate <name> [-c|--clip] [--verbose]
  2ami list [--verbose]
  2ami remove <name> [--verbose]
  2ami rename <old-name> <new-name>
  2ami -h | --help
  2ami --version

Commands:
  add       Add a new key.
  dump      Dump keys informations (without secrets).
  generate  Generate a token from a known key.
  list      List known keys.
  remove    Remove specified key.

Options:
  -h --help             Show this screen.
  --version             Show version.
  --verbose             Enable verbose output.
  --digits=<digits>     Number of token digits.
  --interval=<seconds>  Interval in seconds between token generation.
  -c --clip             Copy result to the clipboard.

Environment variables:
  2AMI_DB    Path to the database where 2FA keys information are stored.
						 Default to $XDG_DATA_HOME/2ami/database.boltdb.
						 For non Linux values of XDG_DATA_HOME see https://github.com/OpenPeeDeeP/xdg
	2AMI_RING	 Name of the keyring/keychain where 2FA secrets will be stored.
						 Default to "login".
`
}

const databaseLocationPerm = 0755

func main() {
	checkAndEnableDebugMode()
	debugPrint("Enabled debug logging...")

	ui := cli.ColoredUi{
		OutputColor: cli.UiColorNone,
		InfoColor:   cli.UiColorBlue,
		ErrorColor:  cli.UiColorRed,
		WarnColor:   cli.UiColorYellow,
		Ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
	}

	viper.SetDefault("db", filepath.Join(xdg.DataHome(), "2ami", "database.boltdb"))
	viper.SetDefault("ring", "login")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("2AMI")

	usage := usage()
	arguments, _ := docopt.ParseDoc(usage)
	debugPrint(fmt.Sprint(arguments))

	databaseLocation, databaseFilename, err := getDatabaseConfigurations()
	if err != nil {
		ui.Error(err.Error())
	}

	debugPrint(fmt.Sprintf("Using database: %s/%s", databaseLocation, databaseFilename))

	err = os.MkdirAll(databaseLocation, databaseLocationPerm)
	if err != nil {
		ui.Error(fmt.Sprintf("Cannot create database location; %s", err))
	}

	storage := NewStorage(databaseLocation, databaseFilename)
	if err = storage.Init(); err != nil {
		ui.Error(fmt.Sprintf("Cannot initialize database; %s", err))
		os.Exit(1)
	}

	verbose = arguments["--verbose"].(bool)

	// deleteAllKeys(storage) //nolint:unused

	if arguments["add"].(bool) {
		name := arguments["<name>"].(string)
		if name == "" {
			ui.Error("argument 'name' cannot be empty")
			os.Exit(1)
		}

		err := add(&ui, storage, name, arguments["--digits"], arguments["--interval"])
		if err != nil {
			ui.Error("An unexpected error occurred. Use DEBUG=true to show logs.")
			debugPrint(fmt.Sprintf("%s", err))
			os.Exit(1)
		}
		os.Exit(0)
	}
	if arguments["dump"].(bool) {
		if arguments["<name>"] == nil {
			errors := dumpAllKeys(storage)
			printErrorsAndExit(errors) // this can exit(1)
			os.Exit(0)
		}
		name := arguments["<name>"].(string)
		err := dumpKey(storage, name)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}
	if arguments["generate"].(bool) {
		name := arguments["<name>"].(string)
		if name == "" {
			ui.Error("argument 'name' cannot be empty")
			os.Exit(1)
		}
		token, err := generate(storage, name)
		if err != nil {
			ui.Error(err.Error())
		}

		if arguments["--clip"].(bool) {
			err = clipboard.WriteAll(token.Value)
			ui.Error(fmt.Sprintf("Cannot copy to clipboard: %s", err))
		} else {
			if verbose {
				ui.Info(fmt.Sprintf("%s ( %d seconds left )\n", token.Value, token.ExpiresIn))
			} else {
				ui.Info(token.Value)
			}
		}
		os.Exit(0)
	}
	if arguments["list"].(bool) {
		errors := list(&ui, storage)
		printErrorsAndExit(errors) // this can exit(1)
		os.Exit(0)
	}
	if arguments["remove"].(bool) {
		name := arguments["<name>"].(string)
		err := remove(&ui, storage, name)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}
	if arguments["rename"].(bool) {
		oldName := arguments["<old-name>"].(string)
		newName := arguments["<new-name>"].(string)
		if oldName == newName {
			ui.Error("old-name and new-name are equal, aborting")
			os.Exit(1)
		}
		err := rename(&ui, storage, oldName, newName)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}
		ui.Info("Key renamed")
		os.Exit(0)
	}
	if arguments["--version"].(bool) {
		ui.Output(version)
		os.Exit(0)
	}

	os.Exit(0)
}

func checkAndEnableDebugMode() {
	_, ok := os.LookupEnv("DEBUG")
	if ok {
		debug = true
	}
}

func add(ui cli.Ui, storage Storage, name string, digits interface{}, interval interface{}) error {
	secret, err := ui.AskSecret(fmt.Sprintf("2fa secret for %s ( will not be printed ): ", name))
	if err != nil {
		return err
	}
	secret = sanitizeSecret(secret)

	if err = isValidBase32(secret); err != nil {
		return fmt.Errorf("secret is not valid: %w", err)
	}

	ring, err := openKeyring()
	if err != nil {
		return fmt.Errorf("cannot open keyring: %w", err)
	}
	key := NewKey(ring, name)
	if digits != nil {
		key.Digits, err = convertStringToInt(digits.(string))
		if err != nil {
			return fmt.Errorf("cannot convert string to int: %w", err)
		}
	}
	if interval != nil {
		key.Interval, err = convertStringToInt(interval.(string))
		if err != nil {
			return fmt.Errorf("cannot convert string to int: %w", err)
		}
	}
	err = key.Secret(secret)
	if err != nil {
		return fmt.Errorf("cannot set secret for key: %w", err)
	}

	debugPrint(fmt.Sprintf("%+v", key))

	marshal, _ := json.Marshal(key)
	debugPrint(string(marshal))

	result, err := storage.AddKey(name, []byte(marshal))
	if err != nil {
		return err
	}
	if !result {
		return errors.New("something went wrong adding key")
	}

	ui.Info("Key successfully added")
	return nil
}

type generated struct {
	Value     string
	ExpiresIn int
}

func generate(storage Storage, name string) (generated, error) {
	ring, err := openKeyring()
	if err != nil {
		return generated{}, fmt.Errorf("cannot open keyring: %w", err)
	}

	key := KeyFromStorage(storage, ring, name)
	return generated{
		Value:     key.GenerateToken(),
		ExpiresIn: key.ExpiresIn(),
	}, nil
}

func list(ui cli.Ui, storage Storage) (errors []error) {
	keys, err := storage.ListKey()
	if err != nil {
		return []error{err}
	}

	for _, v := range keys {
		value, err := storage.GetKey(v)
		if err != nil {
			errors = append(errors, err)
		}
		key := Key{}
		err = json.Unmarshal([]byte(value), &key)
		if err != nil {
			errors = append(errors, err)
		}
		debugPrint(fmt.Sprintf("%+v", key))

		if verbose {
			ui.Output(key.VerboseString())
		} else {
			ui.Output(key.String())
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

//nolint
func deleteAllKeys(storage Storage) {
	keys, err := storage.ListKey()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("keys: %s\n", keys)

	for _, v := range keys {
		err = storage.RemoveKey(v)
		if err != nil {
			fmt.Printf("%+v\n", err)
			// switch err := errors.Cause(err).(type) {
			// default:
			// 	fmt.Printf("%+v\n", err)
			// }
		}
	}
	os.Exit(1)
}

func remove(ui cli.Ui, storage Storage, name string) error {
	ring, err := openKeyring()
	if err != nil {
		return err
	}
	key := KeyFromStorage(storage, ring, name)

	err = key.Delete()
	if err != nil {
		if strings.HasPrefix(err.Error(), "Item not found") {
			ui.Info("Key is not present in keyring, skipping deletion")
		} else {
			return err
		}
	}
	err = storage.RemoveKey(name)
	if err != nil {
		return err
	}
	ui.Info("Key removed")
	return nil
}

func rename(ui cli.Ui, storage Storage, oldName string, newName string) error {
	ring, err := openKeyring()
	if err != nil {
		return err
	}
	key := KeyFromStorage(storage, ring, oldName)
	err = key.Rename(newName)
	if err != nil {
		ui.Error(fmt.Sprintf("Error renaming key %s: %s", oldName, err))
	}
	marshal, _ := json.Marshal(key)
	debugPrint(string(marshal))

	result, err := storage.AddKey(key.Name, []byte(marshal))
	if err != nil {
		return err
	}
	err = storage.RemoveKey(oldName)
	if err != nil {
		ui.Error(fmt.Sprintf("Removal of old key failed: %s", err))
	}
	if !result {
		ui.Error("something went wrong adding key")
		os.Exit(1)
	}
	return nil
}

func getDatabaseConfigurations() (databaseLocation, databaseFilename string, err error) {
	databaseLocation = filepath.Dir(viper.GetString("db"))
	databaseFilename = filepath.Base(viper.GetString("db"))

	return databaseLocation, databaseFilename, nil
}

func sanitizeSecret(data string) string {
	// any newline is not necessary
	data = strings.TrimSuffix(data, "\n")
	// Base32 is always uppercase
	data = strings.ToUpper(data)
	// remove all spaces in the string
	data = strings.ReplaceAll(data, " ", "")
	return data
}

func isValidBase32(data string) error {
	encoder := base32.Encoding{}
	_, err := encoder.DecodeString(data)
	if err != nil {
		return err
	}
	return nil
}
