package post

import (
	"log"

	"github.com/lib/pq"

	"github.com/Feokrat/tkdsnt-posting-app/internal/model"

	"github.com/jmoiron/sqlx"
)

type repository struct {
	db     *sqlx.DB
	logger *log.Logger
}

type Repository interface {
	Create(post model.Post) (int, error)
}

func NewRepository(db *sqlx.DB, logger *log.Logger) Repository {
	return repository{
		db:     db,
		logger: logger,
	}
}

func (r repository) Create(post model.Post) (int, error) {
	query := "INSERT INTO downloaded AS d (filenames, source, posted) VALUES ($1, $2, $3) RETURNING id"

	var id int

	row := r.db.QueryRow(query, pq.Array(post.Filenames), post.SourceUrl, post.Posted)

	if err := row.Scan(&id); err != nil {
		r.logger.Printf("error in db while trying to create post info, error: %s", err)
		return 0, err
	}

	return id, nil
}
