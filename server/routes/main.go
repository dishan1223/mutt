package routes

import (
	"time"

	"github.com/dishan1223/mutt/internal/middleware"
	"github.com/dishan1223/mutt/server/handler"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

var app *fiber.App

func Init(a *fiber.App) {
	app = a

	v1 := app.Group("/api/v1")

	app.Get("/ping", handler.Ping)
	v1.Get("/ping", handler.Ping)

	auth := v1.Group("/auth")
	auth.Post("/signup", handler.SignUpHandler)
	auth.Post("/login", handler.LoginHandler)
	auth.Post("/refresh", handler.RefreshTokenHandler)
	auth.Post("/logout", middleware.AuthRequired, handler.LogoutHandler)
	auth.Get("/me", middleware.AuthRequired, handler.MeHandler)

	projects := v1.Group("/projects", middleware.AuthRequired)
	projects.Post("/", handler.CreateProjectHandler)
	projects.Get("/", handler.ListProjectsHandler)
	projects.Get("/:id", handler.GetProjectHandler)
	projects.Patch("/:id", handler.UpdateProjectHandler)
	projects.Delete("/:id", handler.DeleteProjectHandler)
	projects.Post("/:id/rotate-key", handler.RotateAPIKeyHandler)

	errors := projects.Group("/:id/errors")
	errors.Get("/", handler.ListErrorGroupsHandler)
	errors.Get("/:errorGroupId", handler.GetErrorGroupHandler)
	errors.Patch("/:errorGroupId", handler.UpdateErrorGroupHandler)
	errors.Delete("/:errorGroupId", handler.DeleteErrorGroupHandler)

	// Rate limiter provided by GoFiber.
	// Docs: https://docs.gofiber.io/blog/fiber-v3-rate-limiting-guide
	ingestRateLimit := limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			projectID, _ := c.Locals("projectID").(uint)
			return "ingest:" + string(rune(projectID))
		},
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Ingest rate limit exceeded",
				"retry_after": 60,
			})
		},
	})

	// Ingest endpoint is for the SDKs to send error logs.
	v1.Post("/ingest", middleware.APIKeyAuth, ingestRateLimit, handler.IngestErrorHandler)
}
