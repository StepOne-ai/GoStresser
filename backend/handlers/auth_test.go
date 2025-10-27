package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	db "github.com/StepOne-ai/GoStresser/database"
	"github.com/StepOne-ai/GoStresser/middleware"
	"github.com/StepOne-ai/GoStresser/utils"
	"github.com/gin-gonic/gin"
)

func getTestToken(userID string) string {
    token, _ := utils.GenerateJWT(userID)
    return token
}

func TestRegisterHandler_Success(t *testing.T) {
    db.DB = db.SetupTestDB()
    db.DB.Exec("DELETE FROM users")

    w := httptest.NewRecorder()
    reqBody := `{"email":"register_success@example.com","password":"password123"}`
    req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(reqBody))
    req.Header.Set("Content-Type", "application/json")

	c := gin.New()
	c.POST("/api/v1/auth/register", RegisterHandler)
	c.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    var resp AuthResponse
    err := json.NewDecoder(w.Body).Decode(&resp)
    assert.NoError(t, err)
    assert.Equal(t, "register_success@example.com", resp.User.Email)
    assert.NotEmpty(t, resp.Token)
}

func TestRegisterHandler_DuplicateEmail(t *testing.T) {
    db.DB = db.SetupTestDB()
    db.DB.Exec("DELETE FROM users")

    // First registration
    db.CreateTestUser(db.DB, "duplicate@example.com", "password123")

    // Try to register again
    w := httptest.NewRecorder()
    reqBody := `{"email":"duplicate@example.com","password":"password123"}`
    req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(reqBody))
    req.Header.Set("Content-Type", "application/json")

	c := gin.New()
	c.POST("/api/v1/auth/register", RegisterHandler)
	c.ServeHTTP(w, req)

    assert.Equal(t, http.StatusConflict, w.Code)
    var resp map[string]any
    json.NewDecoder(w.Body).Decode(&resp)
    assert.Equal(t, "User already exists", resp["error"])
}

func TestRegisterHandler_InvalidInput(t *testing.T) {
    db.DB = db.SetupTestDB()
    db.DB.Exec("DELETE FROM users")

    tests := []struct {
        name     string
        body     string
        expected int
    }{
        {"Missing email", `{"password":"password123"}`, http.StatusBadRequest},
        {"Missing password", `{"email":"test@example.com"}`, http.StatusBadRequest},
        {"Invalid email", `{"email":"invalid","password":"password123"}`, http.StatusBadRequest},
        {"Short password", `{"email":"test@example.com","password":"123"}`, http.StatusBadRequest},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := httptest.NewRecorder()
            req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(tt.body))
            req.Header.Set("Content-Type", "application/json")

			c := gin.New()
			c.POST("/api/v1/auth/register", RegisterHandler)
			c.ServeHTTP(w, req)

            assert.Equal(t, tt.expected, w.Code)
        })
    }
}

func TestLoginHandler_Success(t *testing.T) {
    db.DB = db.SetupTestDB()
    db.DB.Exec("DELETE FROM users")

    db.CreateTestUser(db.DB, "login_success@example.com", "password123")

    w := httptest.NewRecorder()
    reqBody := `{"email":"login_success@example.com","password":"password123"}`
    req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(reqBody))
    req.Header.Set("Content-Type", "application/json")

	c := gin.New()
	c.POST("/api/v1/auth/login", LoginHandler)
	c.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var resp AuthResponse
    err := json.NewDecoder(w.Body).Decode(&resp)
    assert.NoError(t, err)
    assert.Equal(t, "login_success@example.com", resp.User.Email)
    assert.NotEmpty(t, resp.Token)
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
    db.DB = db.SetupTestDB()
    db.DB.Exec("DELETE FROM users")

	db.CreateTestUser(db.DB, "login_success@example.com", "password123")

    tests := []struct {
        name     string
        body     string
        expected int
    }{
        {"Wrong password", `{"email":"wrong@example.com","password":"wrongpass"}`, http.StatusUnauthorized},
        {"User not found", `{"email":"notfound@example.com","password":"password123"}`, http.StatusUnauthorized},
        {"Missing email", `{"password":"password123"}`, http.StatusBadRequest},
        {"Missing password", `{"email":"test@example.com"}`, http.StatusBadRequest},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            w := httptest.NewRecorder()
            req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(tt.body))
            req.Header.Set("Content-Type", "application/json")

            c := gin.New()
			c.POST("/api/v1/auth/login", LoginHandler)
			c.ServeHTTP(w, req)

            assert.Equal(t, tt.expected, w.Code)
        })
    }
}

func TestProtectedRoute_Success(t *testing.T) {
    db.DB = db.SetupTestDB()
    db.DB.Exec("DELETE FROM users")

    user := db.CreateTestUser(db.DB, "protected@example.com", "password123")
    token := getTestToken(user.ID)

    protectedHandler := func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "Access granted"})
    }

    r := gin.New()
    r.Use(middleware.AuthMiddleware())
    r.GET("/protected", protectedHandler)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+token)

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}

func TestProtectedRoute_NoToken(t *testing.T) {
    r := gin.New()
    r.Use(middleware.AuthMiddleware())
    r.GET("/protected", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "Access granted"})
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/protected", nil)
    // No Authorization header

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProtectedRoute_InvalidToken(t *testing.T) {
    r := gin.New()
    r.Use(middleware.AuthMiddleware())
    r.GET("/protected", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "Access granted"})
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer invalid.token.here")

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWT_ExpiredToken(t *testing.T) {
    // Generate expired token
    claims := &utils.Claims{
        UserID: "test-user-id",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte("your-secret-key-change-in-production"))

    r := gin.New()
    r.Use(middleware.AuthMiddleware())
    r.GET("/protected", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "Access granted"})
    })

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+tokenString)

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
