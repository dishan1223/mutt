package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/dishan1223/mutt/internal/config"
	"github.com/gofiber/fiber/v3"
)

func RateLimit(maxRequests int, window time.Duration) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := c.Locals("userID").(uint)
		if !ok {
			return c.Next()
		}

		ctx := context.Background()
		key := fmt.Sprintf("ratelimit:%d:%d", userID, time.Now().UnixMilli()/window.Milliseconds())

		pipe := config.RDB.Pipeline()
		incr := pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		_, err := pipe.Exec(ctx)
		if err != nil {
			// fail open
			return c.Next()
		}

		count := incr.Val()
		if count > int64(maxRequests) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Rate limit exceeded",
				"retry_after": window.Seconds(),
			})
		}

		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(int64(maxRequests)-count, 0)))
		return c.Next()
	}
}
