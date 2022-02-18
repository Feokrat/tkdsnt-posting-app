package downloadPost

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
	Create(post model.DownloadedPost) (int, error)
	GetLastUnposted() (model.DownloadedPost, error)
	SetPosted(id uint) error
}

func NewRepository(db *sqlx.DB, logger *log.Logger) Repository {
	return repository{
		db:     db,
		logger: logger,
	}
}

func (r repository) SetPosted(id uint) error {
	query := "UPDATE downloaded SET posted = true WHERE id = $1"

	_, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	return nil
}

func (r repository) GetLastUnposted() (model.DownloadedPost, error) {
	query := "SELECT id, filenames, source FROM downloaded WHERE posted=false LIMIT 1"

	var downloaded model.DownloadedPost
	err := r.db.QueryRow(query).Scan(&downloaded.Id, pq.Array(&downloaded.Filenames), &downloaded.SourceUrl)
	if err != nil {
		return model.DownloadedPost{}, err
	}
	return downloaded, nil
}

func (r repository) Create(post model.DownloadedPost) (int, error) {
	query := "INSERT INTO downloaded AS d (filenames, source, posted) VALUES ($1, $2, $3) RETURNING id"

	var id int

	row := r.db.QueryRow(query, pq.Array(post.Filenames), post.SourceUrl, post.Posted)

	if err := row.Scan(&id); err != nil {
		r.logger.Printf("error in db while trying to create post info, error: %s", err)
		return 0, err
	}

	return id, nil
}
