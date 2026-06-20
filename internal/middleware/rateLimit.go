package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
)

func RateLimit(maxRequests int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, Ok := c.Locals("userID").(uint)
		if !Ok {
			return c.Next()
		}
	}
}
