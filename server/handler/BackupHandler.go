package handler

import (
	"encoding/json"
	"time"

	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/models"
	"github.com/gofiber/fiber/v3"
)

func ExportBackupHandler(c fiber.Ctx) error {
	userId := c.Locals("userID").(uint)

	var projects []models.Project
	if err := config.DB.Where("user_id = ?", userId).Find(&projects).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch projects",
		})
	}

	backup := models.BackupData{
		ExportedAt: time.Now(),
		Projects:   make([]models.BackupProject, 0, len(projects)),
	}

	for _, p := range projects {
		var groups []models.ErrorGroup
		config.DB.Where("project_id = ?", p.ID).Find(&groups)

		bp := models.BackupProject{
			Name:        p.Name,
			ErrorGroups: make([]models.BackupErrorGroup, 0, len(groups)),
		}

		for _, g := range groups {
			var errs []models.Error
			config.DB.Where("error_group_id = ?", g.ID).Find(&errs)

			bg := models.BackupErrorGroup{
				Title:       g.Title,
				Status:      g.Status,
				Fingerprint: g.Fingerprint,
				Count:       g.Count,
				LastSeenAt:  g.LastSeenAt,
				Errors:      make([]models.BackupError, 0, len(errs)),
			}

			for _, e := range errs {
				bg.Errors = append(bg.Errors, models.BackupError{
					Log:        e.Log,
					StackTrace: e.StackTrace,
					Severity:   e.Severity,
					OccurredAt: e.OccurredAt,
				})
			}

			bp.ErrorGroups = append(bp.ErrorGroups, bg)
		}
		backup.Projects = append(backup.Projects, bp)
	}

	data, err := json.Marshal(backup)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to serialize backup data",
		})
	}

	c.Set("Content-Type", "application/json")

	// Backup file is stored with the name: Mutt_Backup.json
	c.Set("content-disposition", "attachment; filename=Mutt_Backup.json")
	return c.Status(fiber.StatusOK).Send(data)
}
