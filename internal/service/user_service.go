package service

import (
	"errors"

	"github.com/caioersb/ecommerce-api/internal/dto"
	"github.com/caioersb/ecommerce-api/internal/model"
	"github.com/caioersb/ecommerce-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// ErrEmailAlreadyExists é retornado ao tentar cadastrar um e-mail já usado.
var ErrEmailAlreadyExists = errors.New("e-mail já cadastrado")

// UserService define o contrato da camada de regra de negócio de usuários.
type UserService interface {
	Create(req dto.CreateUserRequest) (*dto.UserResponse, error)
	List() ([]dto.UserResponse, error)
	GetByID(id uint) (*dto.UserResponse, error)
	Update(id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	Delete(id uint) error
}

// userService é a implementação concreta da regra de negócio.
type userService struct {
	repo repository.UserRepository
}

// NewUserService cria uma nova instância do serviço de usuários.
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Create(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	// Garante que o e-mail ainda não está em uso.
	if _, err := s.repo.FindByEmail(req.Email); err == nil {
		return nil, ErrEmailAlreadyExists
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	role := req.Role
	if role == "" {
		role = model.RoleCustomer
	}

	user := &model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
		Role:     role,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

func (s *userService) List() ([]dto.UserResponse, error) {
	users, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	return dto.ToUserResponseList(users), nil
}

func (s *userService) GetByID(id uint) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	resp := dto.ToUserResponse(user)
	return &resp, nil
}

func (s *userService) Update(id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Se o e-mail mudou, valida se o novo já não pertence a outro usuário.
	if req.Email != nil && *req.Email != user.Email {
		if existing, err := s.repo.FindByEmail(*req.Email); err == nil && existing.ID != user.ID {
			return nil, ErrEmailAlreadyExists
		} else if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
			return nil, err
		}
		user.Email = *req.Email
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Role != nil {
		user.Role = *req.Role
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

func (s *userService) Delete(id uint) error {
	return s.repo.Delete(id)
}
