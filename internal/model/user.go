package model

import (
	"time"

	"gorm.io/gorm"
)

// Role representa o papel do usuário dentro do e-commerce.
type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

// User é a entidade de domínio que representa um usuário do e-commerce.
// As tags do GORM definem as restrições no banco de dados.
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(120);not null" json:"name"`
	Email     string         `gorm:"type:varchar(160);uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"type:varchar(255);not null" json:"-"`
	Role      Role           `gorm:"type:varchar(20);not null;default:customer" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName define explicitamente o nome da tabela.
func (User) TableName() string {
	return "users"
}
