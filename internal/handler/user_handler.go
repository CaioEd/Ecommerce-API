package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/caioersb/ecommerce-api/internal/dto"
	"github.com/caioersb/ecommerce-api/internal/repository"
	"github.com/caioersb/ecommerce-api/internal/service"
	"github.com/gin-gonic/gin"
)

// UserHandler é a camada de apresentação: traduz HTTP <-> serviço.
type UserHandler struct {
	service service.UserService
}

// NewUserHandler cria uma nova instância do handler de usuários.
func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

// Create trata POST /users.
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Create(req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao criar usuário"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// List trata GET /users.
func (h *UserHandler) List(c *gin.Context) {
	users, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar usuários"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// GetByID trata GET /users/:id.
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	user, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao buscar usuário"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Update trata PUT /users/:id.
func (h *UserHandler) Update(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Update(id, req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao atualizar usuário"})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// Delete trata DELETE /users/:id.
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao remover usuário"})
		return
	}

	c.Status(http.StatusNoContent)
}

// parseID converte o parâmetro de rota em um uint válido.
func parseID(param string) (uint, error) {
	id, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
