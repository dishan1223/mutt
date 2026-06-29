package handler

import (
	"errors"

	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/internal/service"
	"github.com/dishan1223/mutt/internal/utils"
	"github.com/dishan1223/mutt/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"gorm.io/gorm"
)

var validate = validator.New()

func SignUpHandler(c fiber.Ctx) error {
	body := new(models.SignupRequest)

	if err := c.Bind().Body(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := validate.Struct(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": utils.FormatValidationErrors(err),
		})
	}

	var existingUser models.User
	res := config.DB.Where("email = ?", body.Email).First(&existingUser)
	if res.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "User with this email already exists",
		})
	}
	if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Error("Database error during email availability check", "error", res.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
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
	}

	result := config.DB.Create(&user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	return sendTokenPair(c, user.ID)
}

func LoginHandler(c fiber.Ctx) error {
	body := new(models.LoginRequest)

	if err := c.Bind().Body(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := validate.Struct(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": utils.FormatValidationErrors(err),
		})
	}

	var user models.User
	res := config.DB.Where("email = ?", body.Email).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid email or password",
			})
		}
		log.Error("Database error during login", "error", res.Error)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	if !service.VerifyPassword(user.Password, body.Password) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	return sendTokenPair(c, user.ID)
}

func LogoutHandler(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken != "" {
		service.DeleteRefreshToken(refreshToken)
	}

	tokenID, ok := c.Locals("tokenID").(string)
	if ok && tokenID != "" {
		service.BlacklistAccessToken(tokenID)
	}

	clearAuthCookies(c)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

func RefreshTokenHandler(c fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		body := new(models.RefreshRequest)
		if err := c.Bind().Body(body); err == nil && body.RefreshToken != "" {
			refreshToken = body.RefreshToken
		}
	}

	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Refresh token required",
		})
	}

	userID, err := service.GetRefreshTokenUserID(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired refresh token",
		})
	}

	service.DeleteRefreshToken(refreshToken)

	return sendTokenPair(c, userID)
}

func MeHandler(c fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var user models.User
	res := config.DB.First(&user, userID)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user": user.ToResponse(),
	})
}

func sendTokenPair(c fiber.Ctx, userID uint) error {
	accessToken, _, err := service.GenerateAccessToken(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate access token",
		})
	}

	refreshToken := service.GenerateRefreshToken()
	if err := service.StoreRefreshToken(refreshToken, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store refresh token",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   15 * 60,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 60 * 60,
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Success",
	})
}

func clearAuthCookies(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		Path:     "/",
		MaxAge:   -1,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
	})
}
