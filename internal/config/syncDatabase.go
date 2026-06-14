package config

import "github.com/dishan1223/mutt/models"

func MustSyncDatabase() {
	if err := DB.AutoMigrate(&models.User{}); err != nil {
		panic(err)
	}
}
