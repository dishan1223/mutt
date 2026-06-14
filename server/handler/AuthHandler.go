package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/internal/service"
	"github.com/dishan1223/mutt/models"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"gorm.io/gorm"
)

func AuthHandler(c fiber.Ctx) error {
	body := new(models.SignupRequest)

	err := c.Bind().Body(body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if user with the same email already exists
	var existingUser models.User
	res := config.DB.Where("email = ?", body.Email).First(&existingUser)
	if res.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "User with this email already exists",
		})
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("Database error during email availability check", "error", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	hashedPassword, err := service.HashPassword(body.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to hash password",
		})
	}

	user := models.User{
		Username: body.Username,
		Email:    body.Email,
		Password: hashedPassword,
		Phone:    body.Phone,
		Plan:     "Free", // Default Plan
	}

	result := config.DB.Create(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	token, err := service.GenerateToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    token,
		HTTPOnly: true,
		// secure should be tuned on for production
		Secure:   false,
		SameSite: "lax",
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 365),
	})

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"token":   token,
	})
}
