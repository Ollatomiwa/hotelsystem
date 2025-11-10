package main

import (

	"golang.org/x/crypto/bcrypt"
)

func hashPassword(password string) (string, error) {
bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
return string(bytes), err
}
