package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/99designs/keyring"
	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"
)

const (
	backupFormat2ami = "2ami"
)

type backup struct {
	Name     string `json:"name"`
	Digits   string `json:"digits"`
	Interval string `json:"interval"`
	Secret   string `json:"secret"`
}

func backupAllKeys(storage Storage, password string) (string, error) {
	ring, err := openKeyring()
	if err != nil {
		return "", err
	}

	keys, err := storage.ListKey()
	if err != nil {
		return "", err
	}

	allKeys := make([]string, 0)

	for _, v := range keys {
		value, err := backupKeyFromRing(storage, ring, v, password)
		if err != nil {
			return "", err
		}
		allKeys = append(allKeys, value)
	}

	return strings.Join(allKeys, "."), nil
}

func restore(storage Storage, input string, password string, format string) error {
	switch format {
	case backupFormat2ami:
		return restore2ami(storage, input, password)
	default:
		return fmt.Errorf("unsupported backup format: %s", format)
	}
}

func restore2ami(storage Storage, input string, password string) error {
	sections := strings.Split(input, ".")

	for _, section := range sections {
		b, err := decryptBackup(section, password)
		if err != nil {
			return err
		}

		var digits interface{}
		if b.Digits != "" {
			digits = b.Digits
		}

		var interval interface{}
		if b.Interval != "" {
			interval = b.Interval
		}

		err = add(storage, b.Name, b.Secret, digits, interval)
		if err != nil {
			return err
		}
	}

	return nil
}

func backupKeyFromRing(storage Storage, ring keyring.Keyring, keyName string, password string) (string, error) {
	debugPrint(fmt.Sprintf("Retrieving and encrypting key '%v' for backup", keyName))

	key := KeyFromStorage(storage, ring, keyName)

	rawSecret, err := key.secret.Value()
	if err != nil {
		return "", err
	}

	secret := base32.StdEncoding.EncodeToString(rawSecret)

	b := backup{
		Name:     key.Name,
		Digits:   strconv.Itoa(key.Digits),
		Interval: strconv.Itoa(key.Interval),
		Secret:   secret,
	}

	return encryptBackup(b, password)
}

func encryptBackup(value backup, password string) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	c, err := aes.NewCipher(deriveKey(password))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	sealed := gcm.Seal(nonce, nonce, encoded, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func decryptBackup(value string, password string) (backup, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return backup{}, err
	}

	c, err := aes.NewCipher(deriveKey(password))
	if err != nil {
		return backup{}, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return backup{}, err
	}

	nonceSize := gcm.NonceSize()
	if len(decoded) < nonceSize {
		return backup{}, errors.New("invalid nonce size, cannot decrypt supplied value")
	}

	nonce, ciphertext := decoded[:nonceSize], decoded[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return backup{}, err
	}

	b := backup{}
	if err := json.Unmarshal(plaintext, &b); err != nil {
		return backup{}, err
	}

	return b, nil
}

func deriveKey(passphrase string) []byte {
	return pbkdf2.Key([]byte(passphrase), nil, 1000, 32, sha256.New)
}
