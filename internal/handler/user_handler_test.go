package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/caioersb/ecommerce-api/internal/dto"
	"github.com/caioersb/ecommerce-api/internal/model"
	"github.com/caioersb/ecommerce-api/internal/repository"
	"github.com/caioersb/ecommerce-api/internal/service"
	"github.com/gin-gonic/gin"
)

// errInesperado representa qualquer falha que o handler não conhece e que,
// portanto, deve virar 500 sem vazar detalhes.
var errInesperado = errors.New("detalhe interno que não pode vazar")

// fakeUserService permite programar a resposta de cada método do serviço.
// Os campos não preenchidos devolvem o zero value.
type fakeUserService struct {
	createFn func(dto.CreateUserRequest) (*dto.UserResponse, error)
	listFn   func() ([]dto.UserResponse, error)
	getFn    func(uint) (*dto.UserResponse, error)
	updateFn func(uint, dto.UpdateUserRequest) (*dto.UserResponse, error)
	deleteFn func(uint) error
}

// Garante em tempo de compilação que o fake continua satisfazendo o contrato real.
var _ service.UserService = (*fakeUserService)(nil)

func (s *fakeUserService) Create(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	if s.createFn == nil {
		return &dto.UserResponse{}, nil
	}
	return s.createFn(req)
}

func (s *fakeUserService) List() ([]dto.UserResponse, error) {
	if s.listFn == nil {
		return nil, nil
	}
	return s.listFn()
}

func (s *fakeUserService) GetByID(id uint) (*dto.UserResponse, error) {
	if s.getFn == nil {
		return &dto.UserResponse{}, nil
	}
	return s.getFn(id)
}

func (s *fakeUserService) Update(id uint, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	if s.updateFn == nil {
		return &dto.UserResponse{}, nil
	}
	return s.updateFn(id, req)
}

func (s *fakeUserService) Delete(id uint) error {
	if s.deleteFn == nil {
		return nil
	}
	return s.deleteFn(id)
}

// newTestRouter monta um Gin mínimo com as mesmas rotas do router.New. Registrar
// aqui evita importar o pacote router, que já depende deste pacote.
func newTestRouter(svc service.UserService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	h := NewUserHandler(svc)
	r := gin.New()
	r.POST("/users", h.Create)
	r.GET("/users", h.List)
	r.GET("/users/:id", h.GetByID)
	r.PUT("/users/:id", h.Update)
	r.DELETE("/users/:id", h.Delete)
	return r
}

func do(t *testing.T, r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCreateReturnsCreated(t *testing.T) {
	r := newTestRouter(&fakeUserService{
		createFn: func(req dto.CreateUserRequest) (*dto.UserResponse, error) {
			return &dto.UserResponse{ID: 1, Name: req.Name, Email: req.Email, Role: model.RoleCustomer}, nil
		},
	})

	rec := do(t, r, http.MethodPost, "/users",
		`{"name":"Caio Silva","email":"caio@example.com","password":"senha123"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, esperava %d", rec.Code, http.StatusCreated)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("resposta não é JSON válido: %v", err)
	}
	if _, ok := body["password"]; ok {
		t.Error("a resposta expôs o campo password")
	}
}

func TestCreateStatusMapping(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		err      error
		want     int
		wantBody string
	}{
		{
			name: "sucesso",
			body: `{"name":"Caio Silva","email":"caio@example.com","password":"senha123"}`,
			want: http.StatusCreated,
		},
		{
			name: "json malformado",
			body: `{"name":`,
			want: http.StatusBadRequest,
		},
		{
			name: "e-mail inválido reprovado no binding",
			body: `{"name":"Caio Silva","email":"nao-e-email","password":"senha123"}`,
			want: http.StatusBadRequest,
		},
		{
			name: "senha curta reprovada no binding",
			body: `{"name":"Caio Silva","email":"caio@example.com","password":"123"}`,
			want: http.StatusBadRequest,
		},
		{
			name: "role fora do oneof",
			body: `{"name":"Caio Silva","email":"caio@example.com","password":"senha123","role":"superadmin"}`,
			want: http.StatusBadRequest,
		},
		{
			name: "e-mail duplicado",
			body: `{"name":"Caio Silva","email":"caio@example.com","password":"senha123"}`,
			err:  service.ErrEmailAlreadyExists,
			want: http.StatusConflict,
		},
		{
			name:     "falha inesperada",
			body:     `{"name":"Caio Silva","email":"caio@example.com","password":"senha123"}`,
			err:      errInesperado,
			want:     http.StatusInternalServerError,
			wantBody: "erro ao criar usuário",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter(&fakeUserService{
				createFn: func(dto.CreateUserRequest) (*dto.UserResponse, error) {
					if tt.err != nil {
						return nil, tt.err
					}
					return &dto.UserResponse{ID: 1}, nil
				},
			})

			rec := do(t, r, http.MethodPost, "/users", tt.body)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, esperava %d (corpo: %s)", rec.Code, tt.want, rec.Body.String())
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("corpo = %s, esperava conter %q", rec.Body.String(), tt.wantBody)
			}
			// O 500 não pode devolver a mensagem original do erro interno.
			if tt.want == http.StatusInternalServerError && strings.Contains(rec.Body.String(), errInesperado.Error()) {
				t.Errorf("a resposta vazou o erro interno: %s", rec.Body.String())
			}
		})
	}
}

func TestGetByIDStatusMapping(t *testing.T) {
	tests := []struct {
		name string
		path string
		err  error
		want int
	}{
		{"sucesso", "/users/1", nil, http.StatusOK},
		{"id não numérico", "/users/abc", nil, http.StatusBadRequest},
		{"id negativo", "/users/-1", nil, http.StatusBadRequest},
		{"usuário inexistente", "/users/1", repository.ErrUserNotFound, http.StatusNotFound},
		{"falha inesperada", "/users/1", errInesperado, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter(&fakeUserService{
				getFn: func(uint) (*dto.UserResponse, error) {
					if tt.err != nil {
						return nil, tt.err
					}
					return &dto.UserResponse{ID: 1}, nil
				},
			})

			rec := do(t, r, http.MethodGet, tt.path, "")

			if rec.Code != tt.want {
				t.Errorf("status = %d, esperava %d (corpo: %s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestUpdateStatusMapping(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
		err  error
		want int
	}{
		{"sucesso", "/users/1", `{"name":"Nome Novo"}`, nil, http.StatusOK},
		{"payload vazio é aceito", "/users/1", `{}`, nil, http.StatusOK},
		{"id não numérico", "/users/abc", `{"name":"Nome Novo"}`, nil, http.StatusBadRequest},
		{"json malformado", "/users/1", `{"name":`, nil, http.StatusBadRequest},
		{"nome curto demais", "/users/1", `{"name":"A"}`, nil, http.StatusBadRequest},
		{"usuário inexistente", "/users/1", `{"name":"Nome Novo"}`, repository.ErrUserNotFound, http.StatusNotFound},
		{"e-mail de outro usuário", "/users/1", `{"email":"outro@example.com"}`, service.ErrEmailAlreadyExists, http.StatusConflict},
		{"falha inesperada", "/users/1", `{"name":"Nome Novo"}`, errInesperado, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter(&fakeUserService{
				updateFn: func(uint, dto.UpdateUserRequest) (*dto.UserResponse, error) {
					if tt.err != nil {
						return nil, tt.err
					}
					return &dto.UserResponse{ID: 1}, nil
				},
			})

			rec := do(t, r, http.MethodPut, tt.path, tt.body)

			if rec.Code != tt.want {
				t.Errorf("status = %d, esperava %d (corpo: %s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}

func TestDeleteStatusMapping(t *testing.T) {
	tests := []struct {
		name string
		path string
		err  error
		want int
	}{
		{"sucesso", "/users/1", nil, http.StatusNoContent},
		{"id não numérico", "/users/abc", nil, http.StatusBadRequest},
		{"usuário inexistente", "/users/1", repository.ErrUserNotFound, http.StatusNotFound},
		{"falha inesperada", "/users/1", errInesperado, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newTestRouter(&fakeUserService{
				deleteFn: func(uint) error { return tt.err },
			})

			rec := do(t, r, http.MethodDelete, tt.path, "")

			if rec.Code != tt.want {
				t.Fatalf("status = %d, esperava %d (corpo: %s)", rec.Code, tt.want, rec.Body.String())
			}
			if tt.want == http.StatusNoContent && rec.Body.Len() != 0 {
				t.Errorf("204 devolveu corpo: %s", rec.Body.String())
			}
		})
	}
}

func TestListStatusMapping(t *testing.T) {
	t.Run("sucesso", func(t *testing.T) {
		r := newTestRouter(&fakeUserService{
			listFn: func() ([]dto.UserResponse, error) {
				return []dto.UserResponse{{ID: 1}, {ID: 2}}, nil
			},
		})

		rec := do(t, r, http.MethodGet, "/users", "")

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, esperava %d", rec.Code, http.StatusOK)
		}

		var users []dto.UserResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &users); err != nil {
			t.Fatalf("resposta não é uma lista JSON: %v", err)
		}
		if len(users) != 2 {
			t.Errorf("len(users) = %d, esperava 2", len(users))
		}
	})

	t.Run("falha inesperada", func(t *testing.T) {
		r := newTestRouter(&fakeUserService{
			listFn: func() ([]dto.UserResponse, error) { return nil, errInesperado },
		})

		rec := do(t, r, http.MethodGet, "/users", "")

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("status = %d, esperava %d", rec.Code, http.StatusInternalServerError)
		}
	})
}
