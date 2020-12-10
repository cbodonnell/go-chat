package main

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

// JWTClaims struct
type JWTClaims struct {
	Username string  `json:"username"`
	Groups   []Group `json:"groups"`
	jwt.StandardClaims
}

func checkClaims(r *http.Request) (*JWTClaims, error) {
	jwtCookie, err := r.Cookie("jwt")
	if err != nil {
		return nil, err
	}
	tokenString := jwtCookie.Value

	claims := &JWTClaims{}
	_, err = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.JWTKey), nil
	})
	if err != nil {
		return claims, err
	}
	return claims, nil
}
