package token

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// DefaultKey 用于测试
const (
	DefaultKey = "jwt-1sNzdiSgnNuxyq2g7xml2JvLArU"
)

type Token struct {
	Account string `json:"acc,omitempty"`
	App     string `json:"app,omitempty"`
	Exp     int64  `json:"exp,omitempty"`
}

var errExpiredToken = errors.New("expired token")

// Valid 验证token时间是否合规
func (t *Token) Valid() error {
	if t.Exp < time.Now().Unix() {
		return errExpiredToken
	}
	return nil
}

// Parse 解析一个token
func Parse(key, tk string) (*Token, error) {
	var token = new(Token)
	_, err := jwt.ParseWithClaims(tk, token, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// Generate 生成一个jwt token
func Generate(key string, token *Token) (string, error) {
	jwtTk := jwt.NewWithClaims(jwt.SigningMethodES256, token)
	return jwtTk.SignedString([]byte(key))
}
