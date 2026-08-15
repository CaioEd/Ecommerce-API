package service

import (
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/caioersb/ecommerce-api/internal/dto"
	"github.com/caioersb/ecommerce-api/internal/model"
	"github.com/caioersb/ecommerce-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// errInfra simula uma falha de infraestrutura vinda do repositório.
var errInfra = errors.New("falha de conexão")

// fakeUserRepository é uma implementação em memória de repository.UserRepository.
// Permite exercitar as regras de negócio sem subir banco nenhum.
type fakeUserRepository struct {
	users  map[uint]model.User
	nextID uint

	// Campos opcionais para forçar falhas em métodos específicos.
	errFindByEmail error
	errCreate      error
	errFindAll     error
	errUpdate      error
}

// Garante em tempo de compilação que o fake continua satisfazendo o contrato real.
var _ repository.UserRepository = (*fakeUserRepository)(nil)

func newFakeUserRepository(seed ...model.User) *fakeUserRepository {
	repo := &fakeUserRepository{users: make(map[uint]model.User), nextID: 1}
	for _, u := range seed {
		if u.ID == 0 {
			u.ID = repo.nextID
		}
		if u.ID >= repo.nextID {
			repo.nextID = u.ID + 1
		}
		repo.users[u.ID] = u
	}
	return repo
}

func (r *fakeUserRepository) Create(user *model.User) error {
	if r.errCreate != nil {
		return r.errCreate
	}
	user.ID = r.nextID
	r.nextID++
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt
	r.users[user.ID] = *user
	return nil
}

func (r *fakeUserRepository) FindAll() ([]model.User, error) {
	if r.errFindAll != nil {
		return nil, r.errFindAll
	}

	users := make([]model.User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}
	// Ordena para o teste não depender da iteração aleatória do map.
	sort.Slice(users, func(i, j int) bool { return users[i].ID < users[j].ID })
	return users, nil
}

func (r *fakeUserRepository) FindByID(id uint) (*model.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return &user, nil
}

func (r *fakeUserRepository) FindByEmail(email string) (*model.User, error) {
	if r.errFindByEmail != nil {
		return nil, r.errFindByEmail
	}
	for _, user := range r.users {
		if user.Email == email {
			return &user, nil
		}
	}
	return nil, repository.ErrUserNotFound
}

func (r *fakeUserRepository) Update(user *model.User) error {
	if r.errUpdate != nil {
		return r.errUpdate
	}
	if _, ok := r.users[user.ID]; !ok {
		return repository.ErrUserNotFound
	}
	user.UpdatedAt = time.Now()
	r.users[user.ID] = *user
	return nil
}

func (r *fakeUserRepository) Delete(id uint) error {
	if _, ok := r.users[id]; !ok {
		return repository.ErrUserNotFound
	}
	delete(r.users, id)
	return nil
}

// ptr devolve o endereço de um valor, para preencher os campos opcionais do
// UpdateUserRequest.
func ptr[T any](v T) *T {
	return &v
}

func TestCreateHashesPassword(t *testing.T) {
	repo := newFakeUserRepository()
	svc := NewUserService(repo)

	resp, err := svc.Create(dto.CreateUserRequest{
		Name:     "Caio Silva",
		Email:    "caio@example.com",
		Password: "senha123",
		Role:     model.RoleAdmin,
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resp.ID == 0 {
		t.Error("esperava um ID gerado pelo repositório")
	}
	if resp.Role != model.RoleAdmin {
		t.Errorf("role = %q, esperava %q", resp.Role, model.RoleAdmin)
	}

	stored, err := repo.FindByID(resp.ID)
	if err != nil {
		t.Fatalf("usuário não foi persistido: %v", err)
	}
	if stored.Password == "senha123" {
		t.Fatal("a senha foi gravada em texto puro")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte("senha123")); err != nil {
		t.Errorf("o hash gravado não confere com a senha original: %v", err)
	}
}

func TestCreateAppliesDefaultRole(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	resp, err := svc.Create(dto.CreateUserRequest{
		Name:     "Sem Role",
		Email:    "sem-role@example.com",
		Password: "senha123",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resp.Role != model.RoleCustomer {
		t.Errorf("role = %q, esperava o padrão %q", resp.Role, model.RoleCustomer)
	}
}

func TestCreateRejectsDuplicateEmail(t *testing.T) {
	repo := newFakeUserRepository(model.User{
		Name:  "Já Existe",
		Email: "caio@example.com",
		Role:  model.RoleCustomer,
	})
	svc := NewUserService(repo)

	_, err := svc.Create(dto.CreateUserRequest{
		Name:     "Outro Caio",
		Email:    "caio@example.com",
		Password: "senha123",
	})

	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("erro = %v, esperava ErrEmailAlreadyExists", err)
	}
}

func TestCreatePropagatesRepositoryFailure(t *testing.T) {
	repo := newFakeUserRepository()
	repo.errFindByEmail = errInfra
	svc := NewUserService(repo)

	_, err := svc.Create(dto.CreateUserRequest{
		Name:     "Caio Silva",
		Email:    "caio@example.com",
		Password: "senha123",
	})

	// Uma falha de infraestrutura não pode virar "e-mail já cadastrado".
	if !errors.Is(err, errInfra) {
		t.Errorf("erro = %v, esperava a falha original do repositório", err)
	}
}

func TestCreatePropagatesPersistenceFailure(t *testing.T) {
	repo := newFakeUserRepository()
	repo.errCreate = errInfra
	svc := NewUserService(repo)

	_, err := svc.Create(dto.CreateUserRequest{
		Name:     "Caio Silva",
		Email:    "caio@example.com",
		Password: "senha123",
	})

	if !errors.Is(err, errInfra) {
		t.Errorf("erro = %v, esperava a falha original do repositório", err)
	}
}

func TestListPropagatesRepositoryFailure(t *testing.T) {
	repo := newFakeUserRepository()
	repo.errFindAll = errInfra
	svc := NewUserService(repo)

	_, err := svc.List()

	if !errors.Is(err, errInfra) {
		t.Errorf("erro = %v, esperava a falha original do repositório", err)
	}
}

func TestListReturnsAllUsers(t *testing.T) {
	svc := NewUserService(newFakeUserRepository(
		model.User{Name: "Primeiro", Email: "um@example.com", Role: model.RoleCustomer},
		model.User{Name: "Segundo", Email: "dois@example.com", Role: model.RoleAdmin},
	))

	users, err := svc.List()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("len(users) = %d, esperava 2", len(users))
	}
	if users[0].Name != "Primeiro" || users[1].Name != "Segundo" {
		t.Errorf("ordem inesperada: %q, %q", users[0].Name, users[1].Name)
	}
}

func TestCreateRejectsPasswordAboveBcryptLimit(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	// O bcrypt recusa senhas com mais de 72 bytes. O binding do DTO já barra
	// isso na borda HTTP, mas o service não pode engolir o erro em silêncio.
	_, err := svc.Create(dto.CreateUserRequest{
		Name:     "Senha Longa",
		Email:    "longa@example.com",
		Password: strings.Repeat("a", 73),
	})

	if err == nil {
		t.Error("esperava erro do bcrypt para senha acima de 72 bytes")
	}
	if errors.Is(err, ErrEmailAlreadyExists) {
		t.Error("o erro do bcrypt foi confundido com conflito de e-mail")
	}
}

func TestGetByIDReturnsUser(t *testing.T) {
	svc := NewUserService(newFakeUserRepository(model.User{
		Name:  "Caio Silva",
		Email: "caio@example.com",
		Role:  model.RoleAdmin,
	}))

	resp, err := svc.GetByID(1)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resp.ID != 1 || resp.Email != "caio@example.com" || resp.Role != model.RoleAdmin {
		t.Errorf("resposta inesperada: %+v", resp)
	}
}

func TestGetByIDReturnsNotFound(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	_, err := svc.GetByID(42)

	if !errors.Is(err, repository.ErrUserNotFound) {
		t.Errorf("erro = %v, esperava ErrUserNotFound", err)
	}
}

func TestUpdateAppliesOnlyProvidedFields(t *testing.T) {
	repo := newFakeUserRepository(model.User{
		Name:  "Nome Antigo",
		Email: "caio@example.com",
		Role:  model.RoleCustomer,
	})
	svc := NewUserService(repo)

	resp, err := svc.Update(1, dto.UpdateUserRequest{
		Name: ptr("Nome Novo"),
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if resp.Name != "Nome Novo" {
		t.Errorf("name = %q, esperava %q", resp.Name, "Nome Novo")
	}
	// Os campos ausentes no payload não podem ser zerados.
	if resp.Email != "caio@example.com" {
		t.Errorf("email = %q, esperava que permanecesse inalterado", resp.Email)
	}
	if resp.Role != model.RoleCustomer {
		t.Errorf("role = %q, esperava que permanecesse inalterada", resp.Role)
	}
}

func TestUpdateAcceptsOwnEmail(t *testing.T) {
	repo := newFakeUserRepository(model.User{
		Name:  "Caio Silva",
		Email: "caio@example.com",
		Role:  model.RoleCustomer,
	})
	svc := NewUserService(repo)

	// Reenviar o próprio e-mail não pode ser tratado como conflito.
	_, err := svc.Update(1, dto.UpdateUserRequest{
		Email: ptr("caio@example.com"),
	})

	if err != nil {
		t.Errorf("erro = %v, esperava nenhum", err)
	}
}

func TestUpdateRejectsEmailOfAnotherUser(t *testing.T) {
	repo := newFakeUserRepository(
		model.User{Name: "Primeiro", Email: "um@example.com", Role: model.RoleCustomer},
		model.User{Name: "Segundo", Email: "dois@example.com", Role: model.RoleCustomer},
	)
	svc := NewUserService(repo)

	_, err := svc.Update(1, dto.UpdateUserRequest{
		Email: ptr("dois@example.com"),
	})

	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("erro = %v, esperava ErrEmailAlreadyExists", err)
	}
}

func TestUpdatePropagatesEmailLookupFailure(t *testing.T) {
	repo := newFakeUserRepository(model.User{
		Name:  "Caio Silva",
		Email: "caio@example.com",
		Role:  model.RoleCustomer,
	})
	repo.errFindByEmail = errInfra
	svc := NewUserService(repo)

	// Trocar o e-mail dispara a checagem de unicidade; se ela falhar por
	// infraestrutura, o erro não pode ser confundido com conflito.
	_, err := svc.Update(1, dto.UpdateUserRequest{Email: ptr("novo@example.com")})

	if !errors.Is(err, errInfra) {
		t.Errorf("erro = %v, esperava a falha original do repositório", err)
	}
}

func TestUpdatePropagatesPersistenceFailure(t *testing.T) {
	repo := newFakeUserRepository(model.User{
		Name:  "Caio Silva",
		Email: "caio@example.com",
		Role:  model.RoleCustomer,
	})
	repo.errUpdate = errInfra
	svc := NewUserService(repo)

	_, err := svc.Update(1, dto.UpdateUserRequest{Name: ptr("Nome Novo")})

	if !errors.Is(err, errInfra) {
		t.Errorf("erro = %v, esperava a falha original do repositório", err)
	}
}

func TestUpdateReturnsNotFound(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	_, err := svc.Update(42, dto.UpdateUserRequest{Name: ptr("Qualquer")})

	if !errors.Is(err, repository.ErrUserNotFound) {
		t.Errorf("erro = %v, esperava ErrUserNotFound", err)
	}
}

func TestDelete(t *testing.T) {
	repo := newFakeUserRepository(model.User{
		Name:  "Caio Silva",
		Email: "caio@example.com",
		Role:  model.RoleCustomer,
	})
	svc := NewUserService(repo)

	if err := svc.Delete(1); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if _, err := repo.FindByID(1); !errors.Is(err, repository.ErrUserNotFound) {
		t.Error("o usuário continua no repositório após o Delete")
	}
}

func TestDeleteReturnsNotFound(t *testing.T) {
	svc := NewUserService(newFakeUserRepository())

	err := svc.Delete(42)

	if !errors.Is(err, repository.ErrUserNotFound) {
		t.Errorf("erro = %v, esperava ErrUserNotFound", err)
	}
}
