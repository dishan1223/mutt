package handler

import (
	"time"

	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/models"
	"github.com/gofiber/fiber/v3"
)

func StatsHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	var projectIDs []uint
	if err := config.DB.Model(&models.Project{}).Where("user_id = ?", userID).Pluck("id", &projectIDs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch projects"})
	}

	stats := models.StatsResponse{
		ByStatus: map[string]int{"critical": 0, "resolved": 0, "recovered": 0},
	}

	stats.TotalProjects = len(projectIDs)

	if len(projectIDs) == 0 {
		return c.JSON(stats)
	}

	if err := config.DB.Model(&models.ErrorGroup{}).Where("project_id IN ?", projectIDs).Select("COUNT(*)").Scan(&stats.TotalErrorGroups).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count error groups"})
	}

	if err := config.DB.Model(&models.Error{}).Where("project_id IN ?", projectIDs).Select("COUNT(*)").Scan(&stats.TotalErrors).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count errors"})
	}

	var statusCounts []struct {
		Status string
		Count  int
	}
	if err := config.DB.Model(&models.ErrorGroup{}).Where("project_id IN ?", projectIDs).Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get status counts"})
	}
	for _, s := range statusCounts {
		stats.ByStatus[s.Status] = s.Count
	}

	since := time.Now().Add(-24 * time.Hour)
	if err := config.DB.Model(&models.Error{}).Where("project_id IN ? AND occurred_at >= ?", projectIDs, since).Select("COUNT(*)").Scan(&stats.ErrorsLast24h).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count recent errors"})
	}

	return c.JSON(stats)
}
