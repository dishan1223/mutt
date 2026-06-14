package main

import (
	"github.com/dishan1223/mutt/consts"
	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/internal/service"
	"github.com/dishan1223/mutt/server/routes"
	"github.com/gofiber/fiber/v3"
)

// Functions that starts with 'Must' will panic if there is an error,
// otherwise it will return the expected value.
func init() {
	config.MustLoadEnv()
	config.MustConnectToDB()
	config.MustSyncDatabase()
	service.MustInitJWT(config.MustGetEnv("JwtSecret"))
}

func main() {
	PORT := consts.PORT
	app := fiber.New()
	routes.Init(app)
	app.Listen(PORT)
}
