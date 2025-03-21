package db

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	DB *gorm.DB
}

func New(backend, params string) (*DB, error) {
	var err error

	dialector, err := initDialector(backend, params)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize physical database: %s", err)
	}

	database, err := gorm.Open(dialector)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %s", err)
	}

	return &DB{DB: database}, nil
}

func (db *DB) DoMigrations() error {
	if err := db.DB.AutoMigrate(&Proxy{}); err != nil {
		return err
	}

	if err := db.DB.AutoMigrate(&Backend{}); err != nil {
		return err
	}

	if err := db.DB.AutoMigrate(&Permission{}); err != nil {
		return err
	}

	if err := db.DB.AutoMigrate(&Token{}); err != nil {
		return err
	}

	if err := db.DB.AutoMigrate(&User{}); err != nil {
		return err
	}

	return nil
}

func initDialector(backend, params string) (gorm.Dialector, error) {
	switch backend {
	case "sqlite":
		if params == "" {
			return nil, fmt.Errorf("sqlite database file not specified")
		}

		return sqlite.Open(params), nil
	case "postgresql":
		if params == "" {
			return nil, fmt.Errorf("postgres DSN not specified")
		}

		return postgres.Open(params), nil
	case "":
		return nil, fmt.Errorf("no database backend specified in environment variables")
	default:
		return nil, fmt.Errorf("unknown database backend specified: %s", os.Getenv(backend))
	}
}
