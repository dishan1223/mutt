package routes

import (
	"github.com/dishan1223/mutt/server/handler"
	"github.com/gofiber/fiber/v3"
)

var app *fiber.App

func Init(a *fiber.App) {
	app = a

	v1 := app.Group("/api/v1")

	app.Get("/ping", handler.Ping)

	v1.Get("/ping", handler.Ping)
}
