package database

import (
	"fmt"

	"github.com/caioersb/ecommerce-api/config"
	"github.com/caioersb/ecommerce-api/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect abre a conexão com o PostgreSQL usando GORM.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir conexão: %w", err)
	}

	return db, nil
}

// Migrate executa o AutoMigrate para todas as entidades do domínio.
// Adicione novas entidades aqui conforme o e-commerce for crescendo
// (ex.: &model.Product{}, &model.Order{}).
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
	)
}
