package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/xiaohu/pingjiao/config"
)

func (u *User) SetSchoolPassword(password string) error {
	key, err := schoolPasswordKey()
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(password), nil)
	//nolint:gocritic
	payload := append(nonce, ciphertext...)
	u.SchoolPasswordEnc = base64.StdEncoding.EncodeToString(payload)
	return nil
}

func (u *User) GetSchoolPassword() (string, error) {
	if u.SchoolPasswordEnc == "" {
		return "", errors.New("school password not set")
	}
	key, err := schoolPasswordKey()
	if err != nil {
		return "", err
	}
	data, err := base64.StdEncoding.DecodeString(u.SchoolPasswordEnc)
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
	if len(data) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted password")
	}
	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func schoolPasswordKey() ([]byte, error) {
	k := config.GetSchoolPasswordEncKey()
	if k == "" {
		return nil, errors.New("SCHOOL_PASSWORD_ENC_KEY is empty")
	}
	if len(k) == 32 {
		return []byte(k), nil
	}
	decoded, err := base64.StdEncoding.DecodeString(k)
	if err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	return nil, errors.New("SCHOOL_PASSWORD_ENC_KEY must be 32 bytes or base64(32 bytes)")
}

