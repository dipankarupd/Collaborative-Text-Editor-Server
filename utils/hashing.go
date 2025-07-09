package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func PerformHash(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		panic(err)
	}

	return string(bytes)
}

func CheckHash(password string, hashedPassword string) (bool, string) {
	fmt.Println(password)
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	check := true
	msg := ""
	if err != nil {
		msg = "incorrect password"
		check = false
	}
	return check, msg
}
