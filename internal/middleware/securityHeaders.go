package middleware

import "github.com/gofiber/fiber/v3"

// This middleware sets several security-related HTTP headers on the response to enhance the security of the application.
func SecurityHeaders(c fiber.Ctx) error {
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set("X-Frame-Options", "DENY")
	c.Set("X-XSS-Protection", "1; mode=block")
	c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	return c.Next()
}
