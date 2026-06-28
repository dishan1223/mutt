package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// You can use any postgres database connection string.
// But for this project, we are using Neon as the database provider.
// For more information about Neon, please visit: https://neon.tech/

func MustConnectToDB() {
	var err error
	// Change the connection string to your own database connection string.
	// If you are not using NeonDB, you can change the connection string to your own
	// database connection string.

	// Check our the postgres driver imported from gorm.io/driver/postgres
	// If you are using a different database, you can change the driver to your own database driver.
	// Docs: https://gorm.io/docs/connecting_to_the_database.html
	dsn := MustGetEnv("NEON_DB_CONNECTION_STRING")

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
}
