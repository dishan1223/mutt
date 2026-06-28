package main

import (
	"strings"

	"github.com/dishan1223/mutt/consts"
	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/internal/service"
	"github.com/dishan1223/mutt/server/routes"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

// Mutt is an open source error tracking and monitoring tool built with Go.
// It provides a simple and efficient way to track, manage, and analyze errors in your
// applications. Mutt is designed to be lightweight, fast, and easy to use, making it an
// ideal choice for developers looking for a reliable error tracking solution.

// This project is built using the GoFiber framework.
// for more information about GoFiber, please visit: https://gofiber.io/

// We are using GORM as the ORM for database operations.
// for more information about GORM, please visit: https://gorm.io/
// Check the internal/config/connectToDB.go file for the database connection details.

func init() {
	config.MustLoadEnv()
	config.MustConnectToDB()
	config.MustSyncDatabase()
	config.MustConnectRedis()
	service.MustInitJWT(config.MustGetEnv("JWT_SECRET"))
}

func main() {
	PORT := consts.GetPort()
	app := fiber.New()

	allowedOrigins := strings.Split(config.MustGetEnv("ALLOWED_ORIGINS"), ",")

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Mutt-Key"},
		AllowCredentials: true,
	}))

	routes.Init(app)
	app.Listen(PORT)
}
