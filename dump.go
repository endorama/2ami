package main

import (
	"encoding/json"
	"fmt"
)

func dumpAllKeys(storage Storage) (errors []error) {
	keys, err := storage.ListKey()
	if err != nil {
		return []error{err}
	}

	allKeys := []Key{}

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
		debugPrint(fmt.Sprintf("%#v", key))

		allKeys = append(allKeys, key)
	}

	if len(errors) > 0 {
		return errors
	}

	marshaledKeys, _ := json.Marshal(allKeys)
	fmt.Println(string(marshaledKeys))

	return nil
}

func dumpKey(storage Storage, name string) (err error) {
	value, err := storage.GetKey(name)
	if err != nil {
		return err
	}

	key := Key{}
	err = json.Unmarshal([]byte(value), &key)
	if err != nil {
		return err
	}
	debugPrint(fmt.Sprintf("%#v", key))

	marshaledKey, _ := json.Marshal(key)
	fmt.Println(string(marshaledKey))

	return nil
}
