package utils

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/hex"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var legacyMD5Pattern = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)

func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func VerifyPassword(storedHash, rawPassword string) (ok bool, needsUpgrade bool) {
	if isBcryptHash(storedHash) {
		err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(rawPassword))
		return err == nil, false
	}

	if !legacyMD5Pattern.MatchString(storedHash) {
		return false, false
	}

	sum := md5.Sum([]byte(rawPassword))
	md5Hash := hex.EncodeToString(sum[:])
	if subtle.ConstantTimeCompare([]byte(strings.ToLower(storedHash)), []byte(md5Hash)) == 1 {
		return true, true
	}
	return false, false
}

func isBcryptHash(hash string) bool {
	return strings.HasPrefix(hash, "$2a$") ||
		strings.HasPrefix(hash, "$2b$") ||
		strings.HasPrefix(hash, "$2y$")
}
