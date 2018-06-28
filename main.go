// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	docopt "github.com/docopt/docopt.go"

	"github.com/atotto/clipboard"
)

const (
	VERSION = "0.1.0-alpha"
)

var (
	debug   = false
	verbose = false
)

func usage() string {
	return `Two factor authenticator agent.

Usage:
  two-factor-authenticator add <name> [--digits=<digits>] [--interval=<seconds>] [--verbose]
  two-factor-authenticator generate <name> [-c|--clip] [--verbose]
  two-factor-authenticator list [--verbose]
  two-factor-authenticator remove <name> [--verbose]
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
  --digits=<digits>     Number of token digits
  --interval=<seconds>  Interval in seconds between token generation
  -c --clip             Copy result to the clipboard`
}

func main() {
	checkAndEnableDebugMode()
	if debug {
		fmt.Println("Enabled debug logging...")
	}

	homeFolder, err := getUserHomeFolder()
	if err != nil {
		log.Fatal(err)
	}

	storage := NewStorage(homeFolder, dbFilename)
	err = storage.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer storage.Close()

	usage := usage()
	arguments, _ := docopt.ParseDoc(usage)
	if debug {
		fmt.Println(arguments)
	}

	verbose = arguments["--verbose"].(bool)

	// deleteAllKeys(storage)

	if arguments["add"].(bool) {
		name := arguments["<name>"].(string)
		if name == "" {
			log.Fatal("name cannot be empty")
		}

		add(storage, name, arguments["--digits"], arguments["--interval"])
		os.Exit(0)
	}
	if arguments["generate"].(bool) {
		name := arguments["<name>"].(string)
		if name == "" {
			log.Fatal("name cannot be empty")
		}
		token := generate(storage, name)

		if arguments["--clip"].(bool) {
			clipboard.WriteAll(token.Value)
		} else {
			if verbose {
				fmt.Printf("%s ( %d seconds left )\n", token.Value, token.ExpiresIn)
			} else {
				fmt.Println(token.Value)
			}
		}
		os.Exit(0)
	}
	if arguments["list"].(bool) {
		list(storage)
		os.Exit(0)
	}
	if arguments["remove"].(bool) {
		name := arguments["<name>"].(string)
		remove(storage, name)
		os.Exit(0)
	}
	if arguments["--version"].(bool) {
		fmt.Println(VERSION)
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

func add(storage Storage, name string, digits interface{}, interval interface{}) {
	fmt.Fprintf(os.Stdout, "2fa secret for %s: ", name)
	secret, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	if debug {
		fmt.Printf("%+v\n", key)
	}

	marshal, _ := json.Marshal(key)
	if debug {
		fmt.Printf("%s\n", marshal)
	}
	result, err := storage.AddKey(name, []byte(marshal))
	if err != nil {
		log.Fatal(err)
	}
	if !result {
		log.Fatal("something went wrong adding key")
	}
	log.Output(2, "added key")
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
		Value:     tokenFormatter("google-authenticator", key.Digits, key.GenerateToken()),
		ExpiresIn: key.ExpiresIn(),
	}
}

func list(storage Storage) {
	keys, err := storage.ListKey()
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range keys {
		value, err := storage.GetKey(v)
		if err != nil {
			fmt.Errorf("%s", err)
		}
		key := Key{}
		err = json.Unmarshal([]byte(value), &key)
		if err != nil {
			fmt.Errorf("%s", err)
		}
		if debug {
			fmt.Printf("%+v\n", key)
		}
		if verbose {
			fmt.Println(key.VerboseString())
		} else {
			fmt.Println(key)
		}
	}
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

func remove(storage Storage, name string) {
	err := storage.RemoveKey(name)
	if err != nil {
		// switch err := errors.Cause(err).(type) {
		// default:
		// 	fmt.Printf("%+v\n", err)
		// }
		fmt.Printf("%+v\n", err)
		log.Fatal(err)
	}
	log.Output(2, "Key removed")
	os.Exit(0)
}
