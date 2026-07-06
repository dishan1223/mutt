package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/dishan1223/mutt/consts"
	"github.com/dishan1223/mutt/internal/config"
	"github.com/dishan1223/mutt/internal/utils"
	"github.com/dishan1223/mutt/models"
	"github.com/gofiber/fiber/v3"
)

// Check consts/backup.go for the constants used in this file
const gzipThreshold = consts.GzipThreshold
const maxImportSize = consts.MaxImportSize
const maxImportRecords = consts.MaxImportRecords

func ExportBackupHandler(c fiber.Ctx) error {
	userId := c.Locals("userID").(uint)

	var projects []models.Project
	if err := config.DB.Where("user_id = ?", userId).Find(&projects).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch projects",
		})
	}

	// Check the utils/GenerateBackupModel.go file for the GenerateBackupModel function
	backup, err := utils.GenerateBackupModel(projects)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate backup",
		})
	}

	data, err := json.Marshal(backup)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to serialize backup data",
		})
	}

	c.Set("Content-Type", "application/json")

	// Stream gzip for large payloads, send small ones directly
	if len(data) > gzipThreshold {
		c.Set("Content-Encoding", "gzip")
		c.Set("Content-Disposition", "attachment; filename=Mutt_Backup.json.gz")

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(data); err != nil {
			gz.Close()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to compress backup",
			})
		}
		gz.Close()
		return c.SendStream(bytes.NewReader(buf.Bytes()))
	}

	c.Set("Content-Disposition", "attachment; filename=Mutt_Backup.json")
	return c.Send(data)
}

func ImportBackupHandler(c fiber.Ctx) error {
	userID := c.Locals("userID").(uint)

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file provided",
		})
	}

	// ponytail: size limit prevents memory abuse
	if file.Size > maxImportSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("File too large, max %dMB", maxImportSize>>20),
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read uploaded file",
		})
	}
	defer src.Close()

	// Handle gzip-compressed imports
	var reader io.Reader = src
	if strings.HasSuffix(file.Filename, ".gz") {
		gz, err := gzip.NewReader(src)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid gzip file",
			})
		}
		defer gz.Close()
		reader = gz
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to read file content",
		})
	}

	var backup models.BackupData
	if err := json.Unmarshal(data, &backup); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid backup format",
		})
	}

	if len(backup.Projects) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Backup contains no projects",
		})
	}

	// ponytail: cap total records to prevent accidental mega-imports
	totalRecords := 0
	for _, p := range backup.Projects {
		totalRecords++
		for _, g := range p.ErrorGroups {
			totalRecords++
			totalRecords += len(g.Errors)
		}
	}
	if totalRecords > maxImportRecords {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Backup too large (max %d records)", maxImportRecords),
		})
	}

	// ponytail: single transaction, rollback on any failure
	tx := config.DB.Begin()
	if tx.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to start transaction",
		})
	}

	imported := map[string]int{"projects": 0, "error_groups": 0, "errors": 0}

	for _, bp := range backup.Projects {
		project := models.Project{
			UserID: userID,
			Name:   bp.Name,
		}
		if err := tx.Create(&project).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to import project",
				"details": bp.Name,
			})
		}
		imported["projects"]++

		for _, bg := range bp.ErrorGroups {
			group := models.ErrorGroup{
				ProjectID:   project.ID,
				Title:       bg.Title,
				Status:      bg.Status,
				Fingerprint: bg.Fingerprint,
				Count:       bg.Count,
				LastSeenAt:  bg.LastSeenAt,
			}
			if err := tx.Create(&group).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":   "Failed to import error group",
					"details": bg.Title,
				})
			}
			imported["error_groups"]++

			for _, be := range bg.Errors {
				errRecord := models.Error{
					ErrorGroupID: group.ID,
					ProjectID:    project.ID,
					Log:          be.Log,
					StackTrace:   be.StackTrace,
					Severity:     be.Severity,
					OccurredAt:   be.OccurredAt,
				}
				if err := tx.Create(&errRecord).Error; err != nil {
					tx.Rollback()
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "Failed to import error record",
					})
				}
				imported["errors"]++
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit import",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Import completed",
		"imported": imported,
	})
}
