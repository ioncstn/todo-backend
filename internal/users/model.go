package users

import (
	"strings"
	"time"
	"todo-app/internal/todos"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID                      uint `gorm:"primarykey"`
	CreatedAt               time.Time
	UpdatedAt               time.Time
	DeletedAt               *time.Time `gorm:"index"`
	FirstName               string     `gorm:"not null" validate:"required"`
	LastName                string     `gorm:"not null" validate:"required"`
	Email                   string     `gorm:"type:varchar(255);uniqueIndex;not null" validate:"required,email"`
	Password                string     `gorm:"not null" validate:"required"`
	IsEmailVerified         bool       `gorm:"default:false"`
	EmailVerificationToken  string     `gorm:"type:varchar(255);index"`
	EmailVerificationExpiry *time.Time
	Todos                   []todos.ToDoItem
}

// BeforeSave : hook before a user is saved
func (u *User) BeforeSave(tx *gorm.DB) (err error) {
	if u.Password != "" {
		hash, err := hashPassword(u.Password)
		if err != nil {
			return nil
		}

		u.Password = hash
	}
	if u.Email != "" {
		u.Email = strings.ToLower(u.Email)
	}

	return
}

// BeforeUpdate : hook before a user is updated
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	updatesMap := tx.Statement.Dest.(map[string]any)

	if tx.Statement.Changed("Password") {
		hash, err := hashPassword(updatesMap["password"].(string))
		if err != nil {
			return nil
		}

		updatesMap["password"] = hash
	}

	if tx.Statement.Changed("Email") {
		updatesMap["email"] = strings.ToLower(updatesMap["email"].(string))
	}

	return
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
