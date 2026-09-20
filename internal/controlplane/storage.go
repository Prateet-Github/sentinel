package controlplane

import (
	"database/sql"

	controlv1 "github.com/Prateet-Github/sentinel/proto"
	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(path string) (*Storage, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) CreateService(name string) error {
	_, err := s.db.Exec(
		`INSERT INTO services (name) VALUES (?)`,
		name,
	)

	return err
}

func (s *Storage) CreateBackend(
	serviceName string,
	backend *controlv1.Backend,
) error {
	_, err := s.db.Exec(
		`INSERT INTO backends (
			name,
			service_name,
			url,
			health_check_path
		) VALUES (?, ?, ?, ?)`,
		backend.GetName(),
		serviceName,
		backend.GetUrl(),
		backend.GetHealthCheckPath(),
	)

	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}
