package handler

import (
	"errors"
	"strconv"

	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/internal/service"
	"github.com/dishan1223/mutt/internal/utils"
	"github.com/dishan1223/mutt/models"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func CreateProjectHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	body := new(models.CreateProjectRequest)
	if err := c.Bind().Body(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := body.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": utils.FormatValidationErrors(err),
		})
	}

	rawKey, hashedKey := service.GenerateAPIKey()
	if rawKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate API key",
		})
	}

	project := models.Project{
		UserID: userID,
		Name:   body.Name,
		APIKey: hashedKey,
		Notify: body.Notify,
		Addr:   body.Addr,
	}

	if err := config.DB.Create(&project).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create project",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":         project.ID,
		"name":       project.Name,
		"api_key":    rawKey,
		"created_at": project.CreatedAt,
	})
}

func ListProjectsHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	page, limit := models.ParsePagination(c.Query("page"), c.Query("limit"))
	offset := (page - 1) * limit

	query := config.DB.Where("user_id = ?", userID)
	if q := c.Query("q"); q != "" {
		if len(q) > 100 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Search query too long (max 100 characters)"})
		}
		query = query.Where("name ILIKE ?", "%"+q+"%")
	}

	var totalCount int64
	query.Model(&models.Project{}).Count(&totalCount)

	var projects []models.Project
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&projects).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch projects",
		})
	}

	responses := make([]models.ProjectResponse, len(projects))
	for i, p := range projects {
		responses[i] = p.ToResponse()
	}

	return c.Status(fiber.StatusOK).JSON(models.PaginatedProjects{
		Projects:   responses,
		Pagination: models.NewPagination(page, limit, totalCount),
	})
}

func GetProjectHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project models.Project
	if err := config.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Project not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"project": project.ToResponse(),
	})
}

func UpdateProjectHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	body := new(models.UpdateProjectRequest)
	if err := c.Bind().Body(body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := body.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": utils.FormatValidationErrors(err),
		})
	}

	var project models.Project
	if err := config.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Project not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	updates := make(map[string]interface{})
	if body.Name != nil {
		updates["name"] = *body.Name
	}
	if body.Notify != nil {
		updates["notify"] = *body.Notify
	}
	if body.Addr != nil {
		updates["addr"] = *body.Addr
	}

	if len(updates) > 0 {
		if err := config.DB.Model(&project).Updates(updates).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update project",
			})
		}
	}
	config.DB.First(&project, project.ID)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"project": project.ToResponse(),
	})
}

func DeleteProjectHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	result := config.DB.Where("id = ? AND user_id = ?", projectID, userID).Delete(&models.Project{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete project",
		})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Project deleted successfully",
	})
}

func RotateAPIKeyHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project models.Project
	if err := config.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Project not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// rawKey is only shown once to the user.
	rawKey, hashedKey := service.GenerateAPIKey()
	if rawKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate new API key",
		})
	}

	if err := config.DB.Model(&project).Update("api_key", hashedKey).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to rotate API key",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"api_key": rawKey,
		"message": "API key rotated. Old key is now invalid. Save this key, it won't be shown again.",
	})
}
