package service

import (
	"strconv"

	"github.com/dishan1223/mutt/consts"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(p string) (string, error) {
	// All env variables that are returned by a function (e.g., GetAPIKeyBytes) are
	// returning a string. So, we need to convert them to int before using them.
	// strconv.Atoi is used to convert the string to int. If the conversion fails, it returns an error.
	HashCost, err := strconv.Atoi(consts.GetHashCost())
	if err != nil {
		return "", err
	}

	hp, err := bcrypt.GenerateFromPassword([]byte(p), HashCost)
	if err != nil {
		return "", err
	}

	return string(hp), nil
}

func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
