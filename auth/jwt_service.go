package auth

import (
	"errors"

	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

var JwtKey = []byte("jwt_secret_key")

type Claims struct {
	UserId  uint64 `json:"userId"`
	Account string `json:"account"`
	jwt.StandardClaims
}

func JwtVerify(token string) (*Claims, error) {
	claims := &Claims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			zap.S().Errorw("JwtVerify, ParseWithClaims err", "err", err)
		}
		return nil, err
	}

	if !tkn.Valid {
		zap.S().Errorw("JwtVerify, token invalid")
		return nil, errors.New("invalid jwt token")
	}

	return claims, nil
}
