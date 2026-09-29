package entity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// NewSecretToken は URL などに埋め込む推測不能なトークンと、その保存用ハッシュを生成します
func NewSecretToken() (token string, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(buf)
	return token, HashSecretToken(token), nil
}

// HashSecretToken は十分な長さの乱数トークンを前提に SHA-256 でハッシュします。同じ値になるため検索に使えます
func HashSecretToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
