package aegis

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParseBackup(t *testing.T) {
	// Test data from Aegis documentation
	testData := `{
		"version": 1,
		"header": {
			"slots": null,
			"params": null
		},
		"db": {
			"version": 1,
			"entries": [
				{
					"type": "totp",
					"uuid": "3ae6f1ad-2e65-4ed2-a953-1ec0dff2386d",
					"name": "TestUser",
					"issuer": "TestService",
					"note": "Test entry",
					"icon": null,
					"icon_mime": null,
					"icon_hash": null,
					"favorite": false,
					"info": {
						"secret": "JBSWY3DPEHPK3PXP",
						"algo": "SHA1",
						"digits": 6,
						"period": 30
					},
					"groups": []
				}
			],
			"groups": []
		}
	}`

	backup, err := ParseBackup([]byte(testData))
	if err != nil {
		t.Fatalf("Failed to parse Aegis backup: %v", err)
	}

	// Verify the structure
	if backup.Version != 1 {
		t.Errorf("Expected version 1, got %d", backup.Version)
	}

	if backup.IsEncrypted() {
		t.Error("Expected backup to not be encrypted")
	}
}

func TestIsEncrypted(t *testing.T) {
	// Test plain backup
	plainBackup := &Backup{
		Version: 1,
		Header: Header{
			Slots:  nil,
			Params: Params{},
		},
	}

	if plainBackup.IsEncrypted() {
		t.Error("Plain backup should not be marked as encrypted")
	}

	// Test encrypted backup
	encryptedBackup := &Backup{
		Version: 1,
		Header: Header{
			Slots: []Slot{
				{
					Type: 1,
					UUID: "test-uuid",
				},
			},
			Params: Params{},
		},
	}

	if !encryptedBackup.IsEncrypted() {
		t.Error("Encrypted backup should be marked as encrypted")
	}
}

func TestParsePlainBackup(t *testing.T) {
	// Test data
	testData := `{
		"version": 1,
		"header": {
			"slots": null,
			"params": null
		},
		"db": {
			"version": 1,
			"entries": [
				{
					"type": "totp",
					"uuid": "3ae6f1ad-2e65-4ed2-a953-1ec0dff2386d",
					"name": "TestUser",
					"issuer": "TestService",
					"note": "Test entry",
					"icon": null,
					"icon_mime": null,
					"icon_hash": null,
					"favorite": false,
					"info": {
						"secret": "JBSWY3DPEHPK3PXP",
						"algo": "SHA1",
						"digits": 6,
						"period": 30
					},
					"groups": []
				}
			],
			"groups": []
		}
	}`

	backup, err := ParseBackup([]byte(testData))
	if err != nil {
		t.Fatalf("Failed to parse backup: %v", err)
	}

	db, err := backup.ParsePlainBackup()
	if err != nil {
		t.Fatalf("Failed to parse plain backup: %v", err)
	}

	// Verify the database structure
	if db.Version != 1 {
		t.Errorf("Expected DB version 1, got %d", db.Version)
	}

	if len(db.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(db.Entries))
	}

	entry := db.Entries[0]
	if entry.Name != "TestUser" {
		t.Errorf("Expected name 'TestUser', got '%s'", entry.Name)
	}

	if entry.Issuer != "TestService" {
		t.Errorf("Expected issuer 'TestService', got '%s'", entry.Issuer)
	}

	if entry.Type != "totp" {
		t.Errorf("Expected type 'totp', got '%s'", entry.Type)
	}

	// Verify the secret
	secret, ok := entry.Info["secret"].(string)
	if !ok {
		t.Fatal("Secret not found or not a string")
	}

	if secret != "JBSWY3DPEHPK3PXP" {
		t.Errorf("Expected secret 'JBSWY3DPEHPK3PXP', got '%s'", secret)
	}

	// Verify digits
	digits, ok := entry.Info["digits"].(float64)
	if !ok {
		t.Fatal("Digits not found or not a number")
	}

	if digits != 6 {
		t.Errorf("Expected digits 6, got %f", digits)
	}

	// Verify period
	period, ok := entry.Info["period"].(float64)
	if !ok {
		t.Fatal("Period not found or not a number")
	}

	if period != 30 {
		t.Errorf("Expected period 30, got %f", period)
	}
}

func TestParsePlainBackupEncrypted(t *testing.T) {
	// Create an encrypted backup
	encryptedBackup := &Backup{
		Version: 1,
		Header: Header{
			Slots: []Slot{
				{
					Type: 1,
					UUID: "test-uuid",
				},
			},
			Params: Params{},
		},
	}

	// Try to parse as plain backup
	_, err := encryptedBackup.ParsePlainBackup()
	if err == nil {
		t.Error("Expected error when parsing encrypted backup as plain")
	}
}

func TestDecryptBackupPlain(t *testing.T) {
	// Create a plain backup
	plainBackup := &Backup{
		Version: 1,
		Header: Header{
			Slots:  nil,
			Params: Params{},
		},
	}

	// Try to decrypt plain backup
	_, err := plainBackup.DecryptBackup("password")
	if err == nil {
		t.Error("Expected error when decrypting plain backup")
	}
}

func TestDeriveKey(t *testing.T) {
	password := "testpassword"
	salt := "27ea9ae53fa2f08a8dcd201615a8229422647b3058f9f36b08f9457e62888be1"
	n := 32768
	r := 8
	p := 1

	key, err := deriveKey(password, salt, n, r, p)
	if err != nil {
		t.Fatalf("Failed to derive key: %v", err)
	}

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}
}

func TestDeriveKeyInvalidSalt(t *testing.T) {
	password := "testpassword"
	invalidSalt := "invalid-hex"
	n := 32768
	r := 8
	p := 1

	_, err := deriveKey(password, invalidSalt, n, r, p)
	if err == nil {
		t.Error("Expected error with invalid salt")
	}
}

func TestDecryptSlotInvalidKey(t *testing.T) {
	key := []byte("test-key-32-bytes-long-key-here")
	invalidEncryptedKey := "invalid-hex"
	nonce := "e9705513ba4951fa7a0608d2"
	tag := "931237af257b83c693ddb8f9a7eddaf0"

	_, err := decryptSlot(key, invalidEncryptedKey, nonce, tag)
	if err == nil {
		t.Error("Expected error with invalid encrypted key")
	}
}

func TestDecryptSlotInvalidNonce(t *testing.T) {
	key := []byte("test-key-32-bytes-long-key-here")
	encryptedKey := "491d44550430ba248986b904b8cffd3a6c5755d176ac877bd11b82c934225017"
	invalidNonce := "invalid-hex"
	tag := "931237af257b83c693ddb8f9a7eddaf0"

	_, err := decryptSlot(key, encryptedKey, invalidNonce, tag)
	if err == nil {
		t.Error("Expected error with invalid nonce")
	}
}

func TestDecryptDBInvalidDatabase(t *testing.T) {
	masterKey := []byte("test-key-32-bytes-long-key-here")
	invalidEncryptedDB := "invalid-base64"
	params := Params{
		Nonce: "e9705513ba4951fa7a0608d2",
		Tag:   "931237af257b83c693ddb8f9a7eddaf0",
	}

	_, err := decryptDB(masterKey, invalidEncryptedDB, params)
	if err == nil {
		t.Error("Expected error with invalid encrypted database")
	}
}

func TestDecryptDBInvalidNonce(t *testing.T) {
	masterKey := []byte("test-key-32-bytes-long-key-here")
	encryptedDB := "dGVzdA==" // base64 encoded "test"
	params := Params{
		Nonce: "invalid-hex",
		Tag:   "931237af257b83c693ddb8f9a7eddaf0",
	}

	_, err := decryptDB(masterKey, encryptedDB, params)
	if err == nil {
		t.Error("Expected error with invalid nonce")
	}
}

// Test with official Aegis test files

func TestOfficialPlainBackup(t *testing.T) {
	// Read the official plain backup file
	data, err := os.ReadFile("testdata/aegis_plain.json")
	if err != nil {
		t.Skipf("Skipping test - could not read test file: %v", err)
	}

	backup, err := ParseBackup(data)
	if err != nil {
		t.Fatalf("Failed to parse official plain backup: %v", err)
	}

	if backup.IsEncrypted() {
		t.Error("Official plain backup should not be marked as encrypted")
	}

	db, err := backup.ParsePlainBackup()
	if err != nil {
		t.Fatalf("Failed to parse official plain backup: %v", err)
	}

	// Verify the structure
	if db.Version != 1 {
		t.Errorf("Expected DB version 1, got %d", db.Version)
	}

	// The official test file has 7 entries
	if len(db.Entries) != 7 {
		t.Errorf("Expected 7 entries, got %d", len(db.Entries))
	}

	// Verify specific entries
	expectedEntries := []struct {
		name   string
		issuer string
		type_  string
		secret string
	}{
		{"Mason", "Deno", "totp", "4SJHB4GSD43FZBAI7C2HLRJGPQ"},
		{"James", "SPDX", "totp", "5OM4WOOGPLQEF6UGN3CPEOOLWU"},
		{"Elijah", "Airbnb", "totp", "7ELGJSGXNCCTV3O6LKJWYFV2RA"},
		{"James", "Issuu", "hotp", "YOOMIXWS5GN6RTBPUFFWKTW5M4"},
		{"Benjamin", "Air Canada", "hotp", "KUVJJOM753IHTNDSZVCNKL7GII"},
		{"Mason", "WWE", "hotp", "5VAML3X35THCEBVRLV24CGBKOY"},
		{"Sophia", "Boeing", "steam", "JRZCL47CMXVOQMNPZR2F7J4RGI"},
	}

	for i, expected := range expectedEntries {
		if i >= len(db.Entries) {
			t.Errorf("Entry %d not found", i)
			continue
		}

		entry := db.Entries[i]
		if entry.Name != expected.name {
			t.Errorf("Entry %d: expected name '%s', got '%s'", i, expected.name, entry.Name)
		}
		if entry.Issuer != expected.issuer {
			t.Errorf("Entry %d: expected issuer '%s', got '%s'", i, expected.issuer, entry.Issuer)
		}
		if entry.Type != expected.type_ {
			t.Errorf("Entry %d: expected type '%s', got '%s'", i, expected.type_, entry.Type)
		}

		secret, ok := entry.Info["secret"].(string)
		if !ok {
			t.Errorf("Entry %d: secret not found or not a string", i)
			continue
		}
		if secret != expected.secret {
			t.Errorf("Entry %d: expected secret '%s', got '%s'", i, expected.secret, secret)
		}
	}
}

func TestOfficialEncryptedBackup(t *testing.T) {
	// Read the official encrypted backup file
	data, err := os.ReadFile("testdata/aegis_encrypted.json")
	if err != nil {
		t.Skipf("Skipping test - could not read test file: %v", err)
	}

	backup, err := ParseBackup(data)
	if err != nil {
		t.Fatalf("Failed to parse official encrypted backup: %v", err)
	}

	if !backup.IsEncrypted() {
		t.Error("Official encrypted backup should be marked as encrypted")
	}

	// The password for the test file is "test" (from the decrypt.py script)
	db, err := backup.DecryptBackup("test")
	if err != nil {
		t.Fatalf("Failed to decrypt official encrypted backup: %v", err)
	}

	// Verify the structure
	if db.Version != 1 {
		t.Errorf("Expected DB version 1, got %d", db.Version)
	}

	// The encrypted backup should contain the same entries as the plain one
	if len(db.Entries) != 7 {
		t.Errorf("Expected 7 entries, got %d", len(db.Entries))
	}

	// Verify that we can extract secrets from the decrypted entries
	for i, entry := range db.Entries {
		secret, ok := entry.Info["secret"].(string)
		if !ok {
			t.Errorf("Entry %d: secret not found or not a string", i)
			continue
		}
		if secret == "" {
			t.Errorf("Entry %d: secret is empty", i)
		}
	}
}

func TestOfficialEncryptedBackupWrongPassword(t *testing.T) {
	// Read the official encrypted backup file
	data, err := os.ReadFile("testdata/aegis_encrypted.json")
	if err != nil {
		t.Skipf("Skipping test - could not read test file: %v", err)
	}

	backup, err := ParseBackup(data)
	if err != nil {
		t.Fatalf("Failed to parse official encrypted backup: %v", err)
	}

	// Try with wrong password
	_, err = backup.DecryptBackup("wrongpassword")
	if err == nil {
		t.Error("Expected error when using wrong password")
	}
}

func TestOfficialBackupRoundTrip(t *testing.T) {
	// Read the official plain backup file
	data, err := os.ReadFile("testdata/aegis_plain.json")
	if err != nil {
		t.Skipf("Skipping test - could not read test file: %v", err)
	}

	backup, err := ParseBackup(data)
	if err != nil {
		t.Fatalf("Failed to parse official plain backup: %v", err)
	}

	db, err := backup.ParsePlainBackup()
	if err != nil {
		t.Fatalf("Failed to parse official plain backup: %v", err)
	}

	// Serialize and deserialize to test round trip
	dbBytes, err := json.Marshal(db)
	if err != nil {
		t.Fatalf("Failed to marshal DB: %v", err)
	}

	var db2 DB
	err = json.Unmarshal(dbBytes, &db2)
	if err != nil {
		t.Fatalf("Failed to unmarshal DB: %v", err)
	}

	// Verify the round trip preserved the data
	if db2.Version != db.Version {
		t.Errorf("Version mismatch: expected %d, got %d", db.Version, db2.Version)
	}

	if len(db2.Entries) != len(db.Entries) {
		t.Errorf("Entries count mismatch: expected %d, got %d", len(db.Entries), len(db2.Entries))
	}
}
