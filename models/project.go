package models

import (
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var validate = validator.New()

type Project struct {
	gorm.Model
	UserID uint   `json:"user_id" gorm:"not null;index"`
	Name   string `json:"name" gorm:"not null;type:varchar(100)"`
	APIKey string `json:"-" gorm:"not null;uniqueIndex;type:varchar(255)"` // SHA-256 hash
	Notify bool   `json:"notify" gorm:"default:false"`
	Addr   string `json:"addr" gorm:"type:varchar(255)"` // Slack webhook URL or email address
}

// Error groups similar errors together (fingerprint-based)
type ErrorGroup struct {
	gorm.Model
	ProjectID   uint      `json:"project_id" gorm:"not null;uniqueIndex:idx_project_fingerprint"`
	Fingerprint string    `json:"fingerprint" gorm:"not null;uniqueIndex:idx_project_fingerprint;type:varchar(64)"`
	Title       string    `json:"title" gorm:"not null;type:varchar(500)"`
	Status      string    `json:"status" gorm:"not null;type:varchar(20);default:'critical'"` // critical|resolved|recovered
	Count       int       `json:"count" gorm:"default:1"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	Notified    bool      `json:"notified" gorm:"default:false"` // whether notification has been sent for this group
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

type CreateProjectRequest struct {
	Name   string `json:"name" validate:"required,min=1,max=100"`
	Notify bool   `json:"notify"`
	Addr   string `json:"addr" validate:"omitempty,max=255"`
}

type UpdateProjectRequest struct {
	Name   *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Notify *bool   `json:"notify,omitempty"`
	Addr   *string `json:"addr,omitempty" validate:"omitempty,max=255"`
}

type ProjectResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Notify    bool      `json:"notify"`
	Addr      string    `json:"addr"`
}

type ProjectWithKeyResponse struct {
	ProjectResponse
	APIKey string `json:"api_key"`
}

func (p *Project) ToResponse() ProjectResponse {
	return ProjectResponse{
		ID:        p.ID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
		Notify:    p.Notify,
		Addr:      p.Addr,
	}
}

type UpdateErrorGroupRequest struct {
	Status string `json:"status" validate:"required,oneof=critical resolved recovered"`
}

type ErrorGroupResponse struct {
	ID         uint      `json:"id"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	Count      int       `json:"count"`
	LastSeenAt time.Time `json:"last_seen_at"`
	CreatedAt  time.Time `json:"created_at"`
	Notified   bool      `json:"notified"`
}

func (eg *ErrorGroup) ToResponse() ErrorGroupResponse {
	return ErrorGroupResponse{
		ID:         eg.ID,
		Title:      eg.Title,
		Status:     eg.Status,
		Count:      eg.Count,
		LastSeenAt: eg.LastSeenAt,
		CreatedAt:  eg.CreatedAt,
		Notified:   eg.Notified,
	}
}

type ErrorResponse struct {
	ID         uint      `json:"id"`
	Log        string    `json:"log"`
	StackTrace string    `json:"stack_trace"`
	Severity   string    `json:"severity"`
	Notified   bool      `json:"notified"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (e *Error) ToResponse() ErrorResponse {
	return ErrorResponse{
		ID:         e.ID,
		Log:        e.Log,
		StackTrace: e.StackTrace,
		Severity:   e.Severity,
		Notified:   e.Notified,
		OccurredAt: e.OccurredAt,
	}
}

type IngestRequest struct {
	Title      string `json:"title" validate:"required,max=500"`
	Log        string `json:"log" validate:"required"`
	StackTrace string `json:"stack_trace"`
	Severity   string `json:"severity" validate:"omitempty,oneof=error warning info"`
}

func (r *IngestRequest) Validate() error {
	return validate.Struct(r)
}

func (r *CreateProjectRequest) Validate() error {
	return validate.Struct(r)
}

func (r *UpdateProjectRequest) Validate() error {
	return validate.Struct(r)
}

func (r *UpdateErrorGroupRequest) Validate() error {
	return validate.Struct(r)
}
