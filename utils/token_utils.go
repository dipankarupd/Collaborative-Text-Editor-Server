package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	db "github.com/dipankarupd/text-editor/db"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

var SECRET_KEY = os.Getenv("SECRET_KEY")

type SignedDetails struct {
	Name   string
	Email  string
	UserId uuid.UUID
	jwt.StandardClaims
}

func GenerateAccessAndRefreshToken(
	userId uuid.UUID,
	name string,
	email string,
) (signedAccessToken string, signedRefreshToken string, err error) {

	claims := &SignedDetails{
		Name:   name,
		Email:  email,
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Minute * time.Duration(15)).Unix(),
		},
	}

	refreshClaims := &SignedDetails{
		Name:   name,
		Email:  email,
		UserId: userId,
		StandardClaims: jwt.StandardClaims{

			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}
	return token, refreshToken, err
}

func UpdateTokens(ctx context.Context, refreshToken string, userId uuid.UUID) error {

	// from db.RedisClient, get the redis client,
	// update the refresh token for the userid.
	return db.RedisClient.Set(ctx, "refresh_token:"+userId.String(), refreshToken, 168*time.Hour).Err()
}

func ValidateToken(tokenString string) (*SignedDetails, string) {
	parsedToken, err := jwt.ParseWithClaims(
		tokenString,
		&SignedDetails{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		// Check if the error is due to token expiration
		if ve, ok := err.(*jwt.ValidationError); ok && ve.Errors == jwt.ValidationErrorExpired {
			return nil, "token expired"
		}
		return nil, "invalid token"
	}

	claims, ok := parsedToken.Claims.(*SignedDetails)
	if !ok || !parsedToken.Valid {
		return nil, "invalid token"
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return nil, "token expired"
	}

	return claims, ""
}

func RefreshTokens(refreshToken string, ctx context.Context) (newAccessToken string, newRefreshToken string, err error) {
	// First, parse the token to extract the claims (without full validation yet)
	claims, msg := ValidateToken(refreshToken)
	if msg != "" {
		return "", "", fmt.Errorf("invalid refresh token")
	}
	fmt.Println(claims)
	// Fetch token from Redis using userId
	fmt.Println("refresh_token:" + claims.UserId.String())
	storedToken, err := db.RedisClient.Get(ctx, "refresh_token:"+claims.UserId.String()).Result()
	if err != nil {
		return "", "", fmt.Errorf("refresh token not found in Redis or Redis error: %v", err)
	}

	// Compare stored token with the provided one
	if storedToken != refreshToken {
		return "", "", fmt.Errorf("refresh token mismatch")
	}

	// Now, we can be confident in the token's legitimacy and proceed
	newAccessToken, newRefreshToken, err = GenerateAccessAndRefreshToken(
		claims.UserId,
		claims.Name,
		claims.Email,
	)

	if err != nil {
		return "", "", fmt.Errorf("failed to generate new tokens: %v", err)
	}

	// Store new refresh token in Redis (replacing the old one)
	err = db.RedisClient.Set(ctx, "refresh_token:"+claims.UserId.String(), newRefreshToken, 168*time.Hour).Err()
	if err != nil {
		return "", "", fmt.Errorf("failed to store in redis")
	}

	return newAccessToken, newRefreshToken, nil
}
