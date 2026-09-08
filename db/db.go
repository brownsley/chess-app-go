package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ResetAndMigrate(db *gorm.DB) error {
	log.Println("[WARNING] Resetting database: Dropping existing tables...")

	_ = db.Migrator().DropTable(&Friendship{}, &User{})

	err := db.AutoMigrate(&User{}, &Friendship{})
	if err != nil {
		return fmt.Errorf("failed to auto-migrate after reset: %w", err)
	}

	log.Println("[SUCCESS] Database successfully reset and migrated!")
	return nil
}

func InitDB(dsn string) (*gorm.DB, error) {
	customLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                                   customLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if os.Getenv("RESET_DB") == "true" {
		if err := ResetAndMigrate(database); err != nil {
			return nil, err
		}
	} else {
		err = database.AutoMigrate(&User{}, &Friendship{})
		if err != nil {
			return nil, fmt.Errorf("failed to auto-migrate tables: %w", err)
		}
		log.Println("Database migration successful: 'users' and 'friendships' tables are ready!")
	}

	DB = database
	return database, nil
}
