package utils

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("your-secret-key-change-in-production")

type Claims struct {
    UserID string `json:"user_id"`
    jwt.RegisteredClaims
}

func GenerateJWT(userID string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)
}

func ValidateJWT(tokenString string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })
    if err != nil {
        return nil, err
    }
    if !token.Valid {
        return nil, jwt.ErrSignatureInvalid
    }
    return claims, nil
}

func SetJWTSecret(key []byte) {
    jwtKey = key
}

func GetJWTSecret() []byte {
    return jwtKey
}