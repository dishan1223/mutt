package service

import (
	"strconv"

	"github.com/dishan1223/mutt/consts"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(p string) (string, error) {
	HashCost, err := strconv.Atoi(consts.HASH_COST)
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
