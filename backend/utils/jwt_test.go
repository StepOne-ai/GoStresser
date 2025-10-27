package utils

import (
    "testing"
    "time"
)

func TestGenerateAndValidateJWT(t *testing.T) {
    token, err := GenerateJWT("user123")
    if err != nil {
        t.Fatal("Failed to generate JWT:", err)
    }

    claims, err := ValidateJWT(token)
    if err != nil {
        t.Fatal("Failed to validate JWT:", err)
    }

    if claims.UserID != "user123" {
        t.Fatal("UserID mismatch")
    }

    if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
        t.Fatal("Token expired immediately")
    }
}