package dto

import (
	"time"

	"github.com/caioersb/ecommerce-api/internal/model"
)

// CreateUserRequest é o payload esperado na criação de um usuário.
type CreateUserRequest struct {
	Name     string     `json:"name" binding:"required,min=2,max=120"`
	Email    string     `json:"email" binding:"required,email"`
	Password string     `json:"password" binding:"required,min=6,max=72"`
	Role     model.Role `json:"role" binding:"omitempty,oneof=customer admin"`
}

// UpdateUserRequest é o payload esperado na atualização de um usuário.
// Todos os campos são opcionais (ponteiros) para permitir atualização parcial.
type UpdateUserRequest struct {
	Name  *string     `json:"name" binding:"omitempty,min=2,max=120"`
	Email *string     `json:"email" binding:"omitempty,email"`
	Role  *model.Role `json:"role" binding:"omitempty,oneof=customer admin"`
}

// UserResponse é a representação pública de um usuário (sem dados sensíveis).
type UserResponse struct {
	ID        uint       `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Role      model.Role `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// ToUserResponse converte a entidade de domínio no DTO de resposta.
func ToUserResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ToUserResponseList converte uma lista de entidades em uma lista de DTOs.
func ToUserResponseList(users []model.User) []UserResponse {
	responses := make([]UserResponse, 0, len(users))
	for i := range users {
		responses = append(responses, ToUserResponse(&users[i]))
	}
	return responses
}
