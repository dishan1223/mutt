package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"not null;type:varchar(70)"`
	Email    string `json:"email" gorm:"type:varchar(255);uniqueIndex;not null" validate:"required,email,max=255"`
	Password string `json:"-" gorm:"not null" validate:"required"`
	Plan     string `json:"plan" gorm:"not null;type:varchar(20);default:'Free'"`

	Phone string `json:"phone" gorm:"type:varchar(20);uniqueIndex;not null" validate:"required,max=20"`

	ProjectID pq.Int64Array `json:"project_id" gorm:"type:integer[];default:'{}'"`
}

type Project struct {
	gorm.Model
	ProjectName string
}
