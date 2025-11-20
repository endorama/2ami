// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/OpenPeeDeeP/xdg"
	docopt "github.com/docopt/docopt.go"
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
  2ami backup <file-path>
  2ami restore <file-path> [--format=<format>]
  2ami -h | --help
  2ami --version

Commands:
  add       Add a new key.
  dump      Dump keys informations (without secrets).
  generate  Generate a token from a known key.
  list      List known keys.
  remove    Remove specified key.
  backup    Backup keys to a specified file (with encryption)
  restore   Restore keys from a specified encrypted file

Options:
  -h --help             Show this screen.
  --version             Show version.
  --verbose             Enable verbose output.
  --digits=<digits>     Number of token digits.
  --interval=<seconds>  Interval in seconds between token generation.
  --format=<format>     Backup format to restore from (2ami, aegis, etc.).
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
	if err := run(); err != nil {
		ui.Error("An unexpected error occurred. Use DEBUG=true to show logs.")
		ui.Error(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
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
		return err
	}

	debugPrint(fmt.Sprintf("Using database: %s/%s", databaseLocation, databaseFilename))

	err = os.MkdirAll(databaseLocation, databaseLocationPerm)
	if err != nil {
		return fmt.Errorf("cannot create database location; %w", err)
	}

	storage := NewStorage(databaseLocation, databaseFilename)
	if err = storage.Init(); err != nil {
		return fmt.Errorf("cannot initialize database: %w", err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			ui.Warn(fmt.Sprintf("cannot close database: %s", err))
		}
	}()

	verbose = arguments["--verbose"].(bool)

	// deleteAllKeys(storage) //nolint:unused

	if arguments["add"].(bool) {
		name := arguments["<name>"].(string)
		if name == "" {
			return fmt.Errorf("argument 'name' cannot be empty")
		}

		err := addWithPrompt(&ui, storage, name, arguments["--digits"], arguments["--interval"])
		if err != nil {
			debugPrint(fmt.Sprintf("%s", err))
			return fmt.Errorf("an unexpected error occurred: %w", err)
		}
		return nil
	}
	if arguments["dump"].(bool) {
		if arguments["<name>"] == nil {
			errs := dumpAllKeys(storage)
			if len(errs) > 0 {
				return fmt.Errorf("cannot dump keys: %w", errors.Join(errs...))
			}
			return nil
		}
		name := arguments["<name>"].(string)
		err := dumpKey(storage, name)
		if err != nil {
			return fmt.Errorf("cannot dump key: %w", err)
		}
		return nil
	}
	if arguments["backup"].(bool) {
		backupPath := arguments["<file-path>"].(string)
		if backupPath == "" {
			return fmt.Errorf("argument 'file-path' cannot be empty")
		}

		password, err := ui.AskSecret("Password for backup file: ")
		if err != nil {
			return fmt.Errorf("cannot read stdin: %w", err)
		}

		data, err := backupAllKeys(storage, password)
		if err != nil {
			return fmt.Errorf("cannot backup: %w", err)
		}

		err = os.WriteFile(backupPath, []byte(data), 0664)
		if err != nil {
			return fmt.Errorf("cannot write backup file: %w", err)
		}
		return nil
	}
	if arguments["restore"].(bool) {
		backupPath := arguments["<file-path>"].(string)
		if backupPath == "" {
			return fmt.Errorf("argument 'file-path' cannot be empty")
		}

		format := backupFormat2ami // default format
		if arguments["--format"] != nil {
			format = arguments["--format"].(string)
		}

		data, err := os.ReadFile(backupPath)
		if err != nil {
			return fmt.Errorf("cannot read backup file: %w", err)
		}

		password, err := ui.AskSecret("Password for backup file: ")
		if err != nil {
			return fmt.Errorf("cannot read stdin: %w", err)
		}

		err = restore(storage, string(data), password, format)
		if err != nil {
			return fmt.Errorf("cannot restore: %w", err)
		}
		return nil
	}
	if arguments["generate"].(bool) {
		name := arguments["<name>"].(string)
		if name == "" {
			return fmt.Errorf("argument 'name' cannot be empty")
		}
		token, err := generate(storage, name)
		if err != nil {
			return fmt.Errorf("cannot generate token: %w", err)
		}

		if arguments["--clip"].(bool) {
			err = clipboard.WriteAll(token.Value)
			if err != nil {
				return fmt.Errorf("cannot write to clipboard: %w", err)
			}
		} else {
			if verbose {
				ui.Info(fmt.Sprintf("%s ( %d seconds left )\n", token.Value, token.ExpiresIn))
			} else {
				ui.Info(token.Value)
			}
		}
		return nil
	}
	if arguments["list"].(bool) {
		errs := list(&ui, storage)
		if len(errs) > 0 {
			return fmt.Errorf("cannot list keys: %w", errors.Join(errs...))
		}
		return nil
	}
	if arguments["remove"].(bool) {
		name := arguments["<name>"].(string)
		err := remove(&ui, storage, name)
		if err != nil {
			return fmt.Errorf("cannot remove key: %w", err)
		}
		return nil
	}
	if arguments["rename"].(bool) {
		oldName := arguments["<old-name>"].(string)
		newName := arguments["<new-name>"].(string)
		if oldName == newName {
			return fmt.Errorf("old-name and new-name are equal, halting execution")
		}
		err := rename(&ui, storage, oldName, newName)
		if err != nil {
			return fmt.Errorf("cannot rename key: %w", err)
		}
		ui.Info("Key renamed")
		return nil
	}
	if arguments["--version"].(bool) {
		ui.Output(version)
		return nil
	}

	return nil
}

func checkAndEnableDebugMode() {
	_, ok := os.LookupEnv("DEBUG")
	if ok {
		debug = true
	}
}

func addWithPrompt(ui cli.Ui, storage Storage, name string, digits interface{}, interval interface{}) error {
	secret, err := ui.AskSecret(fmt.Sprintf("2fa secret for %s ( will not be printed ): ", name))
	if err != nil {
		return err
	}
	secret = sanitizeSecret(secret)
	if err := add(storage, name, secret, digits, interval); err != nil {
		return err
	}

	ui.Info("Key successfully added")
	return nil
}

func add(storage Storage, name string, secret string, digits interface{}, interval interface{}) error {
	if err := isValidBase32(secret); err != nil {
		return fmt.Errorf("secret is not valid: %w", err)
	}

	ring, err := openKeyring()
	if err != nil {
		return fmt.Errorf("cannot open keyring: %w", err)
	}
	key := NewKey(ring, name)
	if digits != nil {
		switch v := digits.(type) {
		case int:
			key.Digits = v
		case string:
			key.Digits, err = convertStringToInt(v)
			if err != nil {
				return fmt.Errorf("cannot convert string to int: %w", err)
			}
		default:
			return fmt.Errorf("unsupported type for digits: %T", digits)
		}
	}
	if interval != nil {
		switch v := interval.(type) {
		case int:
			key.Interval = v
		case string:
			key.Interval, err = convertStringToInt(v)
			if err != nil {
				return fmt.Errorf("cannot convert string to int: %w", err)
			}
		default:
			return fmt.Errorf("unsupported type for interval: %T", interval)
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

// nolint
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
