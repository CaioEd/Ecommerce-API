package main

import (
	"log"

	"github.com/caioersb/ecommerce-api/config"
	"github.com/caioersb/ecommerce-api/internal/database"
	"github.com/caioersb/ecommerce-api/internal/handler"
	"github.com/caioersb/ecommerce-api/internal/repository"
	"github.com/caioersb/ecommerce-api/internal/router"
	"github.com/caioersb/ecommerce-api/internal/service"
)

func main() {
	// Carrega as configurações (variáveis de ambiente).
	cfg := config.Load()

	// Inicializa a conexão com o banco e executa as migrações.
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("erro ao conectar no banco de dados: %v", err)
	}

	if err := database.Migrate(db); err != nil {
		log.Fatalf("erro ao executar migrações: %v", err)
	}

	// Wiring das camadas: repository -> service -> handler.
	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	// Configura o roteador Gin e registra as rotas.
	r := router.New(userHandler)

	addr := ":" + cfg.ServerPort
	log.Printf("servidor iniciado em http://localhost%s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("erro ao iniciar o servidor: %v", err)
	}
}
