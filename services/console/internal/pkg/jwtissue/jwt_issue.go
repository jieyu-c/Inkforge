package jwtissue

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func RandBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func RefreshHash(raw []byte) []byte {
	sum := sha256.Sum256(raw)
	return sum[:]
}

func UAHash(agent string) []byte {
	if agent == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(agent))
	return sum[:]
}

func EncodeRefreshToken(raw []byte) string {
	return base64.RawURLEncoding.EncodeToString(raw)
}

func DecodeRefreshToken(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

// IssueAccess returns a signed JWT (HS256); custom claims are copied into ctx by go-zero Authorize middleware.
func IssueAccess(secret, userID, sid string, ttl time.Duration) (token string, expiresInSec int64, err error) {
	now := time.Now()
	expUnix := now.Add(ttl).Unix()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"sid":     sid,
		"iat":     now.Unix(),
		"exp":     expUnix,
	})
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}
	expiresInSec = int64(ttl / time.Second)
	if expiresInSec < 1 {
		expiresInSec = 1
	}
	return signed, expiresInSec, nil
}
