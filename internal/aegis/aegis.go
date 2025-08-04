// Package aegis provides functionality to parse and decrypt Aegis Authenticator backup files.
//
// Aegis Authenticator is a free, secure and open source 2FA app for Android that supports
// HOTP and TOTP algorithms. This package implements the Aegis backup format specification
// to allow importing 2FA secrets from Aegis backups into other applications.
//
// # Supported Backup Formats
//
// The package supports both encrypted and plain text Aegis backup formats:
//
//   - **Encrypted Backups**: Password-protected backups using AES-256-GCM encryption
//     with scrypt key derivation (N=32768, r=8, p=1)
//   - **Plain Text Backups**: Unencrypted JSON backups for easy import/export
//
// # Supported Entry Types
//
// The package can handle the following OTP entry types from Aegis backups:
//
//   - **TOTP**: Time-based One-Time Password (RFC 6238)
//   - **HOTP**: HMAC-based One-Time Password (RFC 4226)
//   - **Steam**: Steam Guard authenticator tokens
//
// # Usage Example
//
//	// Parse an Aegis backup file
//	data, err := os.ReadFile("aegis_backup.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	backup, err := aegis.ParseBackup(data)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	var db *aegis.DB
//	if backup.IsEncrypted() {
//	    // Decrypt encrypted backup
//	    db, err = backup.DecryptBackup("your_password")
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	} else {
//	    // Parse plain text backup
//	    db, err = backup.ParsePlainBackup()
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
//	// Process the entries
//	for _, entry := range db.Entries {
//	    secret := entry.Info["secret"].(string)
//	    fmt.Printf("Entry: %s (%s) - Secret: %s\n", entry.Name, entry.Issuer, secret)
//	}
//
// # Security Considerations
//
// - The package uses the same cryptographic primitives as Aegis Authenticator
// - Scrypt parameters are fixed to match Aegis implementation (N=32768, r=8, p=1)
// - AES-256-GCM is used for authenticated encryption
// - All cryptographic operations are performed in memory
//
// # Aegis Backup Format
//
// Aegis backups are stored in JSON format with the following structure:
//
//	{
//	    "version": 1,
//	    "header": {
//	        "slots": [...],     // Encryption slots (null for plain text)
//	        "params": {...}     // Encryption parameters (null for plain text)
//	    },
//	    "db": "..."            // Encrypted base64 string or plain JSON object
//	}
//
// For more information about the Aegis backup format, see:
// https://github.com/beemdevelopment/Aegis/blob/master/docs/vault.md
package aegis

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/scrypt"
)

// Backup represents the top-level Aegis backup structure
type Backup struct {
	Version int         `json:"version"`
	Header  Header      `json:"header"`
	DB      interface{} `json:"db"` // Can be string (encrypted) or object (plain)
}

// Header contains the encryption slots and parameters
type Header struct {
	Slots  []Slot `json:"slots"`
	Params Params `json:"params"`
}

// Slot represents an encryption slot (password, biometric, etc.)
type Slot struct {
	Type      int    `json:"type"`
	UUID      string `json:"uuid"`
	Key       string `json:"key"`
	KeyParams Params `json:"key_params"`
	N         int    `json:"n,omitempty"`    // scrypt parameter
	R         int    `json:"r,omitempty"`    // scrypt parameter
	P         int    `json:"p,omitempty"`    // scrypt parameter
	Salt      string `json:"salt,omitempty"` // scrypt salt
}

// Params contains encryption parameters (nonce and tag)
type Params struct {
	Nonce string `json:"nonce"`
	Tag   string `json:"tag"`
}

// DB represents the decrypted database content
type DB struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
	Groups  []Group `json:"groups"`
}

// Entry represents a single TOTP/HOTP entry
type Entry struct {
	Type     string                 `json:"type"`
	UUID     string                 `json:"uuid"`
	Name     string                 `json:"name"`
	Issuer   string                 `json:"issuer"`
	Note     string                 `json:"note,omitempty"`
	Icon     interface{}            `json:"icon"`
	IconMime interface{}            `json:"icon_mime"`
	IconHash interface{}            `json:"icon_hash"`
	Favorite bool                   `json:"favorite"`
	Info     map[string]interface{} `json:"info"`
	Groups   []string               `json:"groups"`
}

// Group represents a group of entries
type Group struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// ParseBackup parses an Aegis backup from JSON
func ParseBackup(data []byte) (*Backup, error) {
	var backup Backup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, fmt.Errorf("failed to parse Aegis backup JSON: %w", err)
	}
	return &backup, nil
}

// IsEncrypted checks if the backup is encrypted
func (b *Backup) IsEncrypted() bool {
	return b.Header.Slots != nil && len(b.Header.Slots) > 0
}

// DecryptBackup decrypts an encrypted Aegis backup
func (b *Backup) DecryptBackup(password string) (*DB, error) {
	if !b.IsEncrypted() {
		return nil, fmt.Errorf("backup is not encrypted")
	}

	// Find a password slot (type 1)
	var passwordSlot *Slot
	for i := range b.Header.Slots {
		if b.Header.Slots[i].Type == 1 { // Password slot
			passwordSlot = &b.Header.Slots[i]
			break
		}
	}

	if passwordSlot == nil {
		return nil, fmt.Errorf("no password slot found in Aegis backup")
	}

	// Derive key using scrypt
	key, err := deriveKey(password, passwordSlot.Salt, passwordSlot.N, passwordSlot.R, passwordSlot.P)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	// Decrypt the master key from the slot
	masterKey, err := decryptSlot(key, passwordSlot.Key, passwordSlot.KeyParams.Nonce, passwordSlot.KeyParams.Tag)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt master key: %w", err)
	}

	// Decrypt the database content
	dbString, ok := b.DB.(string)
	if !ok {
		return nil, fmt.Errorf("encrypted database should be a string")
	}

	decryptedDB, err := decryptDB(masterKey, dbString, b.Header.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt database: %w", err)
	}

	// Parse the decrypted database
	var db DB
	if err := json.Unmarshal(decryptedDB, &db); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted database: %w", err)
	}

	return &db, nil
}

// ParsePlainBackup parses a plain text Aegis backup
func (b *Backup) ParsePlainBackup() (*DB, error) {
	if b.IsEncrypted() {
		return nil, fmt.Errorf("backup is encrypted, use DecryptBackup instead")
	}

	// For plain text backups, the DB field should be an object
	dbBytes, err := json.Marshal(b.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plain database: %w", err)
	}

	var db DB
	if err := json.Unmarshal(dbBytes, &db); err != nil {
		return nil, fmt.Errorf("failed to parse plain database: %w", err)
	}

	return &db, nil
}

// deriveKey derives a key using scrypt with Aegis parameters
func deriveKey(password, salt string, n, r, p int) ([]byte, error) {
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return nil, fmt.Errorf("invalid salt: %w", err)
	}

	// Use scrypt with Aegis parameters
	key, err := scrypt.Key([]byte(password), saltBytes, n, r, p, 32)
	if err != nil {
		return nil, fmt.Errorf("scrypt failed: %w", err)
	}

	return key, nil
}

// decryptSlot decrypts a slot using AES-GCM
func decryptSlot(key []byte, encryptedKey, nonce, tag string) ([]byte, error) {
	// Decode the encrypted key and parameters
	encryptedBytes, err := hex.DecodeString(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("invalid encrypted key: %w", err)
	}

	nonceBytes, err := hex.DecodeString(nonce)
	if err != nil {
		return nil, fmt.Errorf("invalid nonce: %w", err)
	}

	tagBytes, err := hex.DecodeString(tag)
	if err != nil {
		return nil, fmt.Errorf("invalid tag: %w", err)
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Combine ciphertext and tag (tag is appended, not prepended)
	ciphertext := append(encryptedBytes, tagBytes...)

	// Decrypt
	plaintext, err := gcm.Open(nil, nonceBytes, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}

// decryptDB decrypts the database content using AES-GCM
func decryptDB(masterKey []byte, encryptedDB string, params Params) ([]byte, error) {
	// Decode the encrypted database
	encryptedBytes, err := base64.StdEncoding.DecodeString(encryptedDB)
	if err != nil {
		return nil, fmt.Errorf("invalid encrypted database: %w", err)
	}

	// Decode parameters
	nonceBytes, err := hex.DecodeString(params.Nonce)
	if err != nil {
		return nil, fmt.Errorf("invalid nonce: %w", err)
	}

	tagBytes, err := hex.DecodeString(params.Tag)
	if err != nil {
		return nil, fmt.Errorf("invalid tag: %w", err)
	}

	// Create AES-GCM cipher
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Combine ciphertext and tag (tag is appended, not prepended)
	ciphertext := append(encryptedBytes, tagBytes...)

	// Decrypt
	plaintext, err := gcm.Open(nil, nonceBytes, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
