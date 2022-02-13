package database

import (
	"fmt"
	"log"

	"github.com/Feokrat/tkdsnt-posting-app/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewPostgresDB(cfg config.PGConfig, logger *log.Logger) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s "+
		"sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.DBName, cfg.SSLMode))
	if err != nil {
		logger.Printf("failed to open connection to database: %s", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		logger.Printf("failed to connect database: %s", err)
		return nil, err
	}

	return db, err
}

func ClosePostgresDB(db *sqlx.DB) {
	db.Close()
}
