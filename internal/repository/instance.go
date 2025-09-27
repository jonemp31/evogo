package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jonemp31/evogo/internal/models"
)

// InstanceRepository define a interface para operações de banco de dados para instâncias.
type InstanceRepository interface {
	Create(instance *models.Instance) error
	FindByName(name string) (*models.Instance, error)
	FindAll() ([]*models.Instance, error)
	Update(instance *models.Instance) error
	Delete(name string) error
}

type postgresInstanceRepository struct {
	db *sql.DB
}

// NewPostgresInstanceRepository cria uma nova instância do repositório de instâncias.
func NewPostgresInstanceRepository(db *sql.DB) InstanceRepository {
	return &postgresInstanceRepository{db: db}
}

// Create insere uma nova instância no banco de dados.
func (r *postgresInstanceRepository) Create(instance *models.Instance) error {
	instance.ID = uuid.New().String()
	instance.CreatedAt = time.Now()
	instance.UpdatedAt = time.Now()

	query := `INSERT INTO instances (id, name, api_key, webhook_url, status, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(query, instance.ID, instance.Name, instance.ApiKey, instance.WebhookURL, instance.Status, instance.CreatedAt, instance.UpdatedAt)
	return err
}

// FindByName busca uma instância pelo nome.
func (r *postgresInstanceRepository) FindByName(name string) (*models.Instance, error) {
	query := `SELECT id, name, api_key, webhook_url, status, created_at, updated_at FROM instances WHERE name = $1`
	row := r.db.QueryRow(query, name)

	var instance models.Instance
	err := row.Scan(&instance.ID, &instance.Name, &instance.ApiKey, &instance.WebhookURL, &instance.Status, &instance.CreatedAt, &instance.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Retorna nil, nil se não encontrar, para não ser um erro.
		}
		return nil, err
	}
	return &instance, nil
}

// FindAll retorna todas as instâncias do banco de dados.
func (r *postgresInstanceRepository) FindAll() ([]*models.Instance, error) {
	query := `SELECT id, name, api_key, webhook_url, status, created_at, updated_at FROM instances ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []*models.Instance
	for rows.Next() {
		var instance models.Instance
		if err := rows.Scan(&instance.ID, &instance.Name, &instance.ApiKey, &instance.WebhookURL, &instance.Status, &instance.CreatedAt, &instance.UpdatedAt); err != nil {
			return nil, err
		}
		instances = append(instances, &instance)
	}
	return instances, nil
}

// Update atualiza os dados de uma instância existente.
func (r *postgresInstanceRepository) Update(instance *models.Instance) error {
	instance.UpdatedAt = time.Now()
	query := `UPDATE instances SET api_key = $1, webhook_url = $2, status = $3, updated_at = $4 WHERE name = $5`
	_, err := r.db.Exec(query, instance.ApiKey, instance.WebhookURL, instance.Status, instance.UpdatedAt, instance.Name)
	return err
}

// Delete remove uma instância do banco de dados pelo nome.
func (r *postgresInstanceRepository) Delete(name string) error {
	query := `DELETE FROM instances WHERE name = $1`
	_, err := r.db.Exec(query, name)
	return err
}
