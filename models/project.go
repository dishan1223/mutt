package models

import (
	"time"

	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	UserID uint   `json:"user_id" gorm:"not null;index"`
	Name   string `json:"name" gorm:"not null;type:varchar(100)"`
	APIKey string `json:"-" gorm:"not null;uniqueIndex;type:varchar(255)"` // bcrypt hash
	Plan   string `json:"plan" gorm:"not null;type:varchar(20);default:'Free'"`
	Notify bool   `json:"notify" gorm:"default:false"`
}

// Error groups similar errors together (fingerprint-based)
type ErrorGroup struct {
	gorm.Model
	ProjectID   uint      `json:"project_id" gorm:"not null;index"`
	Fingerprint string    `json:"fingerprint" gorm:"not null;index;type:varchar(64)"` // SHA-256 of stack trace
	Title       string    `json:"title" gorm:"not null;type:varchar(500)"`
	Status      string    `json:"status" gorm:"not null;type:varchar(20);default:'critical'"` // critical|resolved|recovered
	Count       int       `json:"count" gorm:"default:1"`
	LastSeenAt  time.Time `json:"last_seen_at"`
}

type Error struct {
	gorm.Model
	ErrorGroupID uint      `json:"error_group_id" gorm:"not null;index"`
	ProjectID    uint      `json:"project_id" gorm:"not null;index"`
	Log          string    `json:"log" gorm:"not null;type:text"`
	StackTrace   string    `json:"stack_trace" gorm:"type:text"`
	Severity     string    `json:"severity" gorm:"type:varchar(20);default:'error'"` // error|warning|info
	Notified     bool      `json:"notified" gorm:"default:false"`
	OccurredAt   time.Time `json:"occurred_at"`
}
