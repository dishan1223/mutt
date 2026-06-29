package handler

import (
	"strconv"
	"time"

	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/internal/service"
	"github.com/dishan1223/mutt/internal/utils"
	"github.com/dishan1223/mutt/models"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"gorm.io/gorm"
)

func IngestErrorHandler(c fiber.Ctx) error {
	projectID := c.Locals("projectID").(uint)

	body := new(models.IngestRequest)
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

	// The `ClampIngestRequestFields` function takes an `IngestRequest` object as input and ensures
	// that the lengths of the `Log` and `StackTrace` fields do not exceed predefined maximum sizes.
	// It retrieves these maximum sizes from environment variables, converts them to integers, and truncates
	// the fields if they exceed the limits.
	// If any error occurs during this process (e.g., invalid size values), it returns an error.
	if err := utils.ClampIngestRequestFields(body); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Server configuration error",
		})
	}

	if body.Severity == "" {
		body.Severity = "error"
	}

	fingerprint := service.ComputeFingerprint(body.StackTrace, body.Title)
	group, err := service.FindOrCreateErrorGroup(projectID, fingerprint, body.Title)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process error group",
		})
	}

	shouldNotify := !group.Notified

	errorRecord := models.Error{
		ErrorGroupID: group.ID,
		ProjectID:    projectID,
		Log:          body.Log,
		StackTrace:   body.StackTrace,
		Severity:     body.Severity,
		Notified:     shouldNotify,
		OccurredAt:   time.Now(),
	}

	if err := config.DB.Create(&errorRecord).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store error",
		})
	}

	// Previews version had some race-condition issues.
	// The following two lines update the error group count and last seen timestamp in a single database operation, reducing the risk of
	config.DB.Model(&group).UpdateColumn("count", gorm.Expr("count + 1"))
	config.DB.Model(&group).Update("last_seen_at", time.Now())

	if shouldNotify {
		config.DB.Model(&group).Update("notified", true)
		log.Infof("[NOTIFICATION] New error group detected — Project ID: %d | Group ID: %d | Title: %s", projectID, group.ID, group.Title)
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message":        "Error ingested",
		"error_group_id": group.ID,
		"error_id":       errorRecord.ID,
	})
}

func ListErrorGroupsHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var project models.Project
	if err := config.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var groups []models.ErrorGroup
	if err := config.DB.Where("project_id = ?", projectID).Order("last_seen_at DESC").Find(&groups).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch error groups",
		})
	}

	responses := make([]models.ErrorGroupResponse, len(groups))
	for i, g := range groups {
		responses[i] = g.ToResponse()
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error_groups": responses,
	})
}

func GetErrorGroupHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	errorGroupID, err := strconv.Atoi(c.Params("errorGroupId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid error group ID",
		})
	}

	var project models.Project
	if err := config.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var group models.ErrorGroup
	if err := config.DB.Where("id = ? AND project_id = ?", errorGroupID, projectID).First(&group).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Error group not found",
		})
	}

	var errors []models.Error
	if err := config.DB.Where("error_group_id = ?", errorGroupID).Order("occurred_at DESC").Limit(50).Find(&errors).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch errors",
		})
	}

	errorResponses := make([]models.ErrorResponse, len(errors))
	for i, e := range errors {
		errorResponses[i] = e.ToResponse()
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error_group": group.ToResponse(),
		"errors":      errorResponses,
	})
}

func UpdateErrorGroupHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	errorGroupID, err := strconv.Atoi(c.Params("errorGroupId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid error group ID",
		})
	}

	body := new(models.UpdateErrorGroupRequest)
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
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	var group models.ErrorGroup
	if err := config.DB.Where("id = ? AND project_id = ?", errorGroupID, projectID).First(&group).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Error group not found",
		})
	}

	if err := config.DB.Model(&group).Update("status", body.Status).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update error group",
		})
	}

	group.Status = body.Status
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"error_group": group.ToResponse(),
	})
}

func DeleteErrorGroupHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)
	projectID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	errorGroupID, err := strconv.Atoi(c.Params("errorGroupId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid error group ID",
		})
	}

	var project models.Project
	if err := config.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	result := config.DB.Where("id = ? AND project_id = ?", errorGroupID, projectID).Delete(&models.ErrorGroup{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete error group",
		})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Error group not found",
		})
	}

	config.DB.Where("error_group_id = ?", errorGroupID).Delete(&models.Error{})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Error group deleted successfully",
	})
}
