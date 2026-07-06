package utils

import (
	"time"

	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/models"
)

func GenerateBackupModel(projects []models.Project) (models.BackupData, error) {
	// backup is a single model that contains all the data to be exported
	// Check the models/backup.go file for the structure of the BackupData model
	backup := models.BackupData{
		ExportedAt: time.Now(),
		Projects:   make([]models.BackupProject, 0, len(projects)),
	}

	for _, p := range projects {
		var groups []models.ErrorGroup
		result := config.DB.Where("project_id = ?", p.ID).Find(&groups)
		if result.Error != nil {
			return models.BackupData{}, result.Error
		}

		bp := models.BackupProject{
			Name:        p.Name,
			ErrorGroups: make([]models.BackupErrorGroup, 0, len(groups)),
		}

		for _, g := range groups {
			var errs []models.Error
			result = config.DB.Where("error_group_id = ?", g.ID).Find(&errs)
			if result.Error != nil {
				return models.BackupData{}, result.Error
			}

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

	return backup, nil
}
