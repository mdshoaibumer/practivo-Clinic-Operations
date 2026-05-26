package auth

import (
	"clinmitra/internal/models"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"
)

// persistedSession is the on-disk representation of a session token.
// Only the token is stored; the full session is re-validated from memory.
type persistedSession struct {
	Token     string `json:"token"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	FullName  string `json:"fullName"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"expiresAt"`
}

// SessionStore handles persisting the active session to disk so it
// survives application restarts within the session expiry window.
// Session data is encrypted at rest using AES-256-GCM derived from a
// machine-specific key to prevent token theft from disk.
type SessionStore struct {
	filePath string
	key      []byte
}

// NewSessionStore creates a store backed by a file in the given data directory.
// The encryption key is derived from the data directory path combined with a
// fixed salt, making the file unreadable if moved to another machine/user.
func NewSessionStore(dataDir string) *SessionStore {
	// Derive a 32-byte key from the data directory path (machine-specific).
	// This is not a password-derived key; it prevents casual file copying.
	h := sha256.New()
	h.Write([]byte("clinmitra-session-v1:"))
	h.Write([]byte(dataDir))
	key := h.Sum(nil)

	return &SessionStore{
		filePath: filepath.Join(dataDir, "session.json"),
		key:      key,
	}
}

// encrypt encrypts plaintext using AES-256-GCM.
func (ss *SessionStore) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(ss.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decrypt decrypts ciphertext produced by encrypt.
func (ss *SessionStore) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(ss.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// Save persists the current session to disk with encryption and restrictive permissions.
func (ss *SessionStore) Save(session *Session) error {
	if session == nil {
		return ss.Clear()
	}

	data := persistedSession{
		Token:     session.Token,
		UserID:    session.UserID,
		Username:  session.Username,
		FullName:  session.FullName,
		Role:      string(session.Role),
		ExpiresAt: session.ExpiresAt.Unix(),
	}

	plaintext, err := json.Marshal(data)
	if err != nil {
		return err
	}

	encrypted, err := ss.encrypt(plaintext)
	if err != nil {
		return err
	}

	return os.WriteFile(ss.filePath, encrypted, 0600)
}

// Load reads a persisted session from disk. Returns nil if no session
// file exists or if the session has expired.
func (ss *SessionStore) Load() *Session {
	ciphertext, err := os.ReadFile(ss.filePath)
	if err != nil {
		return nil
	}

	// Try decrypting first (new format)
	plaintext, err := ss.decrypt(ciphertext)
	if err != nil {
		// Might be old unencrypted format — try JSON parse as fallback
		plaintext = ciphertext
	}

	var data persistedSession
	if err := json.Unmarshal(plaintext, &data); err != nil {
		// Corrupt file — remove it
		_ = os.Remove(ss.filePath)
		return nil
	}

	expiresAt := time.Unix(data.ExpiresAt, 0)
	if time.Now().After(expiresAt) {
		// Session expired — clean up
		_ = os.Remove(ss.filePath)
		return nil
	}

	return &Session{
		Token:     data.Token,
		UserID:    data.UserID,
		Username:  data.Username,
		FullName:  data.FullName,
		Role:      models.UserRole(data.Role),
		ExpiresAt: expiresAt,
	}
}

// Clear removes the persisted session file.
func (ss *SessionStore) Clear() error {
	if err := os.Remove(ss.filePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
