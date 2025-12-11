package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func IsLoggedIn(r *http.Request) User {
	// Placeholder logic for checking if a user is logged in
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return User{}
	}
	return newUserFromCookie(cookie.Value)
}

type User struct {
	Id       string
	Username string
}

func (u User) IsEmpty() bool {
	return u.Username == ""
}

// LoginUserWithDB authenticates a user against the database
func LoginUserWithDB(db *pgxpool.Pool, username, password string) (string, error) {
	user, err := AuthenticateUser(db, username, password)
	if err != nil {
		return "", err
	}

	cookie, err := CreateUserCookie(*user)
	if err != nil {
		return "", fmt.Errorf("failed to create cookie: %w", err)
	}

	return cookie, nil
}

// CreateUserCookie creates an encrypted cookie string from a User object
func CreateUserCookie(user User) (string, error) {
	jsonData, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	return EncryptCookie(string(jsonData))
}

func newUserFromCookie(cookie string) User {
	decryptedData, err := DecryptCookie(cookie)
	if err != nil {
		return User{}
	}

	var user User
	err = json.Unmarshal([]byte(decryptedData), &user)
	if err != nil {
		return User{}
	}

	return user
}

type Login struct {
	Username string
	Password string
}

func EncryptCookie(data string) (string, error) {
	key := getSecretKey()
	if key == nil {
		return "", fmt.Errorf("failed to get secret key")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func DecryptCookie(encryptedData string) (string, error) {
	key := getSecretKey()
	if key == nil {
		return "", fmt.Errorf("failed to get secret key")
	}

	data, err := base64.URLEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func getSecretKey() []byte {

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Printf("SECRET_KEY not found in environment")
		return nil
	}

	// Ensure the key is exactly 32 bytes for AES-256
	key := []byte(secretKey)
	if len(key) < 32 {
		// Pad with zeros if too short
		padded := make([]byte, 32)
		copy(padded, key)
		return padded
	}
	if len(key) > 32 {
		// Truncate if too long
		return key[:32]
	}
	return key
}
