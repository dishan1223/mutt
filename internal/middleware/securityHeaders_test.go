package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/models"
	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v3"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	var err error
	config.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	config.DB.Use(otelgorm.NewPlugin())
	config.DB.AutoMigrate(&models.User{}, &models.Project{}, models.ErrorGroup{}, models.Error{})
}

func TestSecurityHeaders(t *testing.T) {
	app := fiber.New()
	app.Use(SecurityHeaders)
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing X-Content-Type-Options header")
	}
	if resp.Header.Get("X-Frame-Options") != "DENY" {
		t.Fatal("missing X-Frame-Options header")
	}
	if resp.Header.Get("X-XSS-Protection") != "1; mode=block" {
		t.Fatal("missing X-XSS-Protection header")
	}
	if resp.Header.Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Fatal("missing Referrer-Policy header")
	}
}
