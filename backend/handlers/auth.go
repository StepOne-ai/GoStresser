package handlers

import (
	"net/http"

	db "github.com/StepOne-ai/GoStresser/database"
	"github.com/StepOne-ai/GoStresser/models"
	"github.com/StepOne-ai/GoStresser/utils"
	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
    Token string      `json:"token"`
    User  models.User `json:"user"`
}

func RegisterHandler(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var existingUser models.User
    if err := db.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
        return
    }

    user := models.User{
        Email:    req.Email,
        Password: req.Password,
    }
    if err := db.DB.Create(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
        return
    }

    token, err := utils.GenerateJWT(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusCreated, AuthResponse{
        Token: token,
        User:  user,
    })
}

func LoginHandler(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var user models.User
    if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    if !utils.CheckPasswordHash(req.Password, user.Password) {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
        return
    }

    token, err := utils.GenerateJWT(user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
        return
    }

    c.JSON(http.StatusOK, AuthResponse{
        Token: token,
        User:  user,
    })
}