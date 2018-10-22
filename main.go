// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	docopt "github.com/docopt/docopt.go"

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
	return `Two factor authenticator agent.

Usage:
  two-factor-authenticator add <name> [--digits=<digits>] [--interval=<seconds>] [--db=<db-path>] [--verbose]
  two-factor-authenticator generate <name> [-c|--clip] [--db=<db-path>] [--verbose]
  two-factor-authenticator list [--db=<db-path>] [--verbose]
  two-factor-authenticator remove <name> [--db=<db-path>] [--verbose]
  two-factor-authenticator -h | --help
  two-factor-authenticator --version

Commands:
  add       Add a new key.
  generate  Generate a token from a known key.
  list      List known keys.
  remove    Remove specified key.

Options:
  -h --help             Show this screen.
  --version             Show version.
  --verbose             Enable verbose output.
  --db=<db-path>        Path to the keys database.
  --digits=<digits>     Number of token digits.
  --interval=<seconds>  Interval in seconds between token generation.
  -c --clip             Copy result to the clipboard.
`
}

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

	usage := usage()
	arguments, _ := docopt.ParseDoc(usage)
	debugPrint(fmt.Sprint(arguments))

	databaseLocation, databaseFilename := getDatabaseConfigurations(arguments["--db"])

	debugPrint(fmt.Sprintf("Using database: %s/%s", databaseLocation, databaseFilename))

	os.MkdirAll(databaseLocation, 0755)
	storage := NewStorage(databaseLocation, databaseFilename)
	if err := storage.Init(); err != nil {
		ui.Error(fmt.Sprintf("Cannot initialize database; %s", err))
		os.Exit(1)
	}

	verbose = arguments["--verbose"].(bool)

	// deleteAllKeys(storage)

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
	if arguments["generate"].(bool) {
		name := arguments["<name>"].(string)
		if name == "" {
			ui.Error("argument 'name' cannot be empty")
			os.Exit(1)
		}
		token := generate(storage, name)

		if arguments["--clip"].(bool) {
			clipboard.WriteAll(token.Value)
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
		if errors != nil {
			for _, element := range errors {
				ui.Error(element.Error())
			}
			os.Exit(1)
		}
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
	secret = strings.TrimSuffix(secret, "\n")

	key := NewKey(name)
	if digits != nil {
		key.Digits = convertStringToInt(digits.(string))
	}
	if interval != nil {
		key.Interval = convertStringToInt(interval.(string))
	}
	err = key.Secret(secret)
	if err != nil {
		return err
	}

	debugPrint(fmt.Sprintf("%+v", key))

	marshal, _ := json.Marshal(key)
	debugPrint(fmt.Sprintf("%s", marshal))

	result, err := storage.AddKey(name, []byte(marshal))
	if err != nil {
		return err
	}
	if !result {
		ui.Error("something went wrong adding key")
		os.Exit(1)
	}
	ui.Info("Key successfully added")
	return nil
}

func generate(storage Storage, name string) struct {
	Value     string
	ExpiresIn int
} {
	key := KeyFromStorage(storage, name)
	return struct {
		Value     string
		ExpiresIn int
	}{
		Value:     key.GenerateToken(),
		ExpiresIn: key.ExpiresIn(),
	}
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
	key := KeyFromStorage(storage, name)

	err := key.Delete()
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

func getDatabaseConfigurations(dbArgument interface{}) (databaseLocation, databaseFilename string) {
	userHome, err := getUserHomeFolder()
	if err != nil {
		log.Fatal(err)
	}

	databaseLocation = userHome
	databaseFilename = ".2fa.db"

	// getting XDG_CONFIG_HOME would be good
	switch runtime.GOOS {
	case "linux":
		databaseLocation = path.Join(userHome, ".config", "two-factor-authenticator")
		databaseFilename = "2fa.db"
	case "darwin":
		databaseLocation = path.Join(userHome, ".config", "two-factor-authenticator")
		databaseFilename = "2fa.db"
	}

	if dbArgument != nil {
		db := dbArgument.(string)
		databaseLocation = filepath.Dir(db)
		databaseFilename = filepath.Base(db)
	}

	return databaseLocation, databaseFilename
}
