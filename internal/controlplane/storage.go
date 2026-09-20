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

func (s *Storage) AddBackend(
	serviceName string,
	backend *controlv1.Backend,
) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO services (name)
         VALUES (?)
         ON CONFLICT(name) DO NOTHING`,
		serviceName,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
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
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Storage) LoadServices() ([]*controlv1.Service, error) {
	rows, err := s.db.Query(`
		SELECT name
		FROM services
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := make([]*controlv1.Service, 0)

	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		services = append(services, &controlv1.Service{
			Name: name,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return services, nil
}

func (s *Storage) LoadBackends() (
	map[string][]*controlv1.Backend,
	error,
) {
	rows, err := s.db.Query(`
		SELECT
			service_name,
			name,
			url,
			health_check_path
		FROM backends
		ORDER BY service_name, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	backends := make(map[string][]*controlv1.Backend)

	for rows.Next() {
		var (
			serviceName     string
			name            string
			url             string
			healthCheckPath string
		)

		if err := rows.Scan(
			&serviceName,
			&name,
			&url,
			&healthCheckPath,
		); err != nil {
			return nil, err
		}

		backends[serviceName] = append(
			backends[serviceName],
			&controlv1.Backend{
				Name:            name,
				Url:             url,
				HealthCheckPath: healthCheckPath,
			},
		)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return backends, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
