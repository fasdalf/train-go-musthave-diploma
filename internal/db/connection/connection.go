package connection

import (
	"context"
	"fmt"
	"github.com/fasdalf/train-go-musthave-diploma/internal/db/entity"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection(ctx context.Context, dbDSN string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dbDSN), &gorm.Config{Logger: slogGorm.New()})

	if err != nil {
		return nil, fmt.Errorf("failed to connect DB: %w", err)
	}
	if err = MigrateDB(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to migrate DB: %w", err)
	}

	return db, nil
}

func MigrateDB(ctx context.Context, db *gorm.DB) error {
	tx := db.WithContext(ctx).Begin()

	entities := []interface{}{
		entity.User{},
		entity.Order{},
		entity.Withdrawal{},
	}

	for _, v := range entities {
		if err := tx.AutoMigrate(&v); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration entity error: %t, %w", v, err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("migration commit error: %w", err)
	}

	return nil
}
