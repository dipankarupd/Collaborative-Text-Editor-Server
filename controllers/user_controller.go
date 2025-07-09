package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	database "github.com/dipankarupd/text-editor/db"
	"github.com/dipankarupd/text-editor/models"
	"github.com/dipankarupd/text-editor/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// db := InitControllers()
func RegisterUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		type RegisterRequest struct {
			Name     string `json:"name" validate:"required,min=2,max=30"`
			Email    string `json:"email" validate:"email,required"`
			Password string `json:"password" validate:"required,min=6"`
		}

		// parse the request body:
		var requestBody RegisterRequest

		if err := ctx.BindJSON(&requestBody); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Request Body"})
			return
		}

		validate := validator.New()
		if err := validate.Struct(requestBody); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}
		// check in the database if the username and email already exists, handle
		var existingUser models.User
		res := db.Where("email = ? OR name = ?", requestBody.Email, requestBody.Name).First(&existingUser)

		if res.Error != nil {
			if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
			// If we get here, it means the user doesn't exist (RecordNotFound)
		} else {
			// User exists
			if existingUser.Email == requestBody.Email {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
			} else {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
			}
			return
		}

		// hash password and add all the required fields
		hashedPassword := utils.PerformHash(requestBody.Password)

		// create a user model
		user := models.User{
			ID:           uuid.New(),
			Name:         requestBody.Name,
			Email:        requestBody.Email,
			PasswordHash: &hashedPassword,
			Provider:     "local",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// store the user in database
		if err := db.Create(&user).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// create the access and refresh token
		accessToken, refreshToken, err := utils.GenerateAccessAndRefreshToken(user.ID, user.Name, user.Email)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating tokens"})
			return
		}
		// store the refresh token to the db
		// hashedRefreshToken := utils.PerformHash(refreshToken)
		err = database.RedisClient.Set(ctx, "refresh_token:"+user.ID.String(), refreshToken, 168*time.Hour).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store the token"})
			return
		}

		// send the response to the user
		response := models.AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         user,
		}
		ctx.JSON(http.StatusCreated, response)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var users []models.User

		result := db.Find(&users)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching the users"})
			return
		}

		ctx.JSON(http.StatusOK, users)
	}
}

func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		userId := ctx.Param("id")

		var user models.User

		result := db.First(&user, "id = ?", userId)

		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching the users"})
			return
		}

		ctx.JSON(http.StatusOK, user)
	}
}

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		type LoginRequest struct {
			Email    string `json:"email" validate:"email,required"`
			Password string `json:"password" validate:"required,min=6"`
		}

		
		var loginRequest LoginRequest
		if err := ctx.BindJSON(&loginRequest); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login Credentials"})
			return
		}

		validate := validator.New()
		if err := validate.Struct(loginRequest); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed. Invalid Credentials"})
			return
		}

		print(loginRequest.Email)
		// validate if user already exists
		var user models.User
		res := db.Where("email = ?", loginRequest.Email).First(&user)

		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not registered. Please register to continue."})
			return
		}
		if res.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		fmt.Println(loginRequest.Password)
		// compare the password hash
		validPassword, err := utils.CheckHash(loginRequest.Password, *user.PasswordHash)
		if !validPassword {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
			return
		}

		// generate new tokens
		accessToken, refreshToken, _ := utils.GenerateAccessAndRefreshToken(user.ID, user.Name, user.Email)

		// update the tokens and store the new refresh token on redis
		if err := utils.UpdateTokens(ctx, refreshToken, user.ID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update refresh token"})
			return
		}
		// return the response
		response := models.AuthResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         user,
		}
		ctx.JSON(http.StatusOK, response)
	}
}

func RefreshHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		refreshToken := ctx.Request.Header.Get("refresh-token")

		if refreshToken == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "no refresh token provided"})
			return
		}
		newAccessToken, newRefreshToken, err := utils.RefreshTokens(refreshToken, context.Background())

		// fmt.Println(err.Error())
		if err != nil {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"error": "invalid token"},
			)

			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"access_token":  newAccessToken,
			"refresh_token": newRefreshToken,
		})
	}
}


func GetLoggedInUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userIdVal, exist := ctx.Get("userid")

		if !exist {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
			return
		}

		userId, ok := userIdVal.(uuid.UUID)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			return
		} 

		var user models.User 
		
		if err := db.First(&user, "id = ?", userId).Error ; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		ctx.JSON(http.StatusOK, user);

	}
}

func Logout() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userIdVal, exist := ctx.Get("userid")

		if !exist {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id not found"})
			return
		}

		userId, ok := userIdVal.(uuid.UUID)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			return
		}
		
		key := "refresh_token:" + userId.String()
		err := database.RedisClient.Del(ctx, key).Err()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to logout"})
        	return
		}

		ctx.JSON(http.StatusOK, gin.H{"success": "logout success"})
	}
}