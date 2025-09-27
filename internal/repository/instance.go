package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/evolution-api/evolution-go/internal/models"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// InstanceRepository handles instance data operations
type InstanceRepository struct {
	db *sql.DB
}

// NewInstanceRepository creates a new instance repository
func NewInstanceRepository(db *sql.DB) *InstanceRepository {
	return &InstanceRepository{db: db}
}

// Create creates a new instance
func (r *InstanceRepository) Create(instance *models.Instance) error {
	query := `
		INSERT INTO instances (
			id, name, token, status, owner_jid, profile_name, profile_pic_url,
			number, business_id, integration, webhook_url, webhook_events,
			webhook_headers, webhook_base64, settings, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	now := time.Now()
	instance.ID = uuid.New().String()
	instance.CreatedAt = now
	instance.UpdatedAt = now

	_, err := r.db.Exec(query,
		instance.ID,
		instance.Name,
		instance.Token,
		instance.Status,
		instance.OwnerJID,
		instance.ProfileName,
		instance.ProfilePicURL,
		instance.Number,
		instance.BusinessID,
		instance.Integration,
		instance.WebhookURL,
		fmt.Sprintf("{%s}", fmt.Sprintf(`"%s"`, instance.WebhookEvents)),
		fmt.Sprintf("{%s}", fmt.Sprintf(`"%s"`, instance.WebhookHeaders)),
		instance.WebhookBase64,
		fmt.Sprintf(`{
			"rejectCall": %t,
			"msgCall": "%s",
			"groupsIgnore": %t,
			"alwaysOnline": %t,
			"readMessages": %t,
			"readStatus": %t,
			"syncFullHistory": %t,
			"wavoipToken": "%s",
			"autoReadMessages": %t,
			"readDelay": %d
		}`, instance.Settings.RejectCall, instance.Settings.MsgCall, instance.Settings.GroupsIgnore,
			instance.Settings.AlwaysOnline, instance.Settings.ReadMessages, instance.Settings.ReadStatus,
			instance.Settings.SyncFullHistory, instance.Settings.WAVoipToken, instance.Settings.AutoReadMessages,
			instance.Settings.ReadDelay),
		instance.CreatedAt,
		instance.UpdatedAt,
	)

	if err != nil {
		zap.L().Error("Failed to create instance", zap.Error(err))
		return err
	}

	return nil
}

// GetByID retrieves an instance by ID
func (r *InstanceRepository) GetByID(id string) (*models.Instance, error) {
	query := `
		SELECT id, name, token, status, owner_jid, profile_name, profile_pic_url,
			   number, business_id, integration, webhook_url, webhook_events,
			   webhook_headers, webhook_base64, settings, created_at, updated_at, last_seen
		FROM instances WHERE id = $1
	`

	instance := &models.Instance{}
	err := r.db.QueryRow(query, id).Scan(
		&instance.ID,
		&instance.Name,
		&instance.Token,
		&instance.Status,
		&instance.OwnerJID,
		&instance.ProfileName,
		&instance.ProfilePicURL,
		&instance.Number,
		&instance.BusinessID,
		&instance.Integration,
		&instance.WebhookURL,
		&instance.WebhookEvents,
		&instance.WebhookHeaders,
		&instance.WebhookBase64,
		&instance.Settings,
		&instance.CreatedAt,
		&instance.UpdatedAt,
		&instance.LastSeen,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("instance not found")
		}
		zap.L().Error("Failed to get instance by ID", zap.Error(err))
		return nil, err
	}

	return instance, nil
}

// GetByName retrieves an instance by name
func (r *InstanceRepository) GetByName(name string) (*models.Instance, error) {
	query := `
		SELECT id, name, token, status, owner_jid, profile_name, profile_pic_url,
			   number, business_id, integration, webhook_url, webhook_events,
			   webhook_headers, webhook_base64, settings, created_at, updated_at, last_seen
		FROM instances WHERE name = $1
	`

	instance := &models.Instance{}
	err := r.db.QueryRow(query, name).Scan(
		&instance.ID,
		&instance.Name,
		&instance.Token,
		&instance.Status,
		&instance.OwnerJID,
		&instance.ProfileName,
		&instance.ProfilePicURL,
		&instance.Number,
		&instance.BusinessID,
		&instance.Integration,
		&instance.WebhookURL,
		&instance.WebhookEvents,
		&instance.WebhookHeaders,
		&instance.WebhookBase64,
		&instance.Settings,
		&instance.CreatedAt,
		&instance.UpdatedAt,
		&instance.LastSeen,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("instance not found")
		}
		zap.L().Error("Failed to get instance by name", zap.Error(err))
		return nil, err
	}

	return instance, nil
}

// GetAll retrieves all instances
func (r *InstanceRepository) GetAll() ([]*models.Instance, error) {
	query := `
		SELECT id, name, token, status, owner_jid, profile_name, profile_pic_url,
			   number, business_id, integration, webhook_url, webhook_events,
			   webhook_headers, webhook_base64, settings, created_at, updated_at, last_seen
		FROM instances ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		zap.L().Error("Failed to query instances", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var instances []*models.Instance
	for rows.Next() {
		instance := &models.Instance{}
		err := rows.Scan(
			&instance.ID,
			&instance.Name,
			&instance.Token,
			&instance.Status,
			&instance.OwnerJID,
			&instance.ProfileName,
			&instance.ProfilePicURL,
			&instance.Number,
			&instance.BusinessID,
			&instance.Integration,
			&instance.WebhookURL,
			&instance.WebhookEvents,
			&instance.WebhookHeaders,
			&instance.WebhookBase64,
			&instance.Settings,
			&instance.CreatedAt,
			&instance.UpdatedAt,
			&instance.LastSeen,
		)
		if err != nil {
			zap.L().Error("Failed to scan instance", zap.Error(err))
			return nil, err
		}
		instances = append(instances, instance)
	}

	return instances, nil
}

// Update updates an instance
func (r *InstanceRepository) Update(instance *models.Instance) error {
	query := `
		UPDATE instances SET
			name = $2, token = $3, status = $4, owner_jid = $5, profile_name = $6,
			profile_pic_url = $7, number = $8, business_id = $9, integration = $10,
			webhook_url = $11, webhook_events = $12, webhook_headers = $13,
			webhook_base64 = $14, settings = $15, updated_at = $16, last_seen = $17
		WHERE id = $1
	`

	instance.UpdatedAt = time.Now()

	_, err := r.db.Exec(query,
		instance.ID,
		instance.Name,
		instance.Token,
		instance.Status,
		instance.OwnerJID,
		instance.ProfileName,
		instance.ProfilePicURL,
		instance.Number,
		instance.BusinessID,
		instance.Integration,
		instance.WebhookURL,
		fmt.Sprintf("{%s}", fmt.Sprintf(`"%s"`, instance.WebhookEvents)),
		fmt.Sprintf("{%s}", fmt.Sprintf(`"%s"`, instance.WebhookHeaders)),
		instance.WebhookBase64,
		fmt.Sprintf(`{
			"rejectCall": %t,
			"msgCall": "%s",
			"groupsIgnore": %t,
			"alwaysOnline": %t,
			"readMessages": %t,
			"readStatus": %t,
			"syncFullHistory": %t,
			"wavoipToken": "%s",
			"autoReadMessages": %t,
			"readDelay": %d
		}`, instance.Settings.RejectCall, instance.Settings.MsgCall, instance.Settings.GroupsIgnore,
			instance.Settings.AlwaysOnline, instance.Settings.ReadMessages, instance.Settings.ReadStatus,
			instance.Settings.SyncFullHistory, instance.Settings.WAVoipToken, instance.Settings.AutoReadMessages,
			instance.Settings.ReadDelay),
		instance.UpdatedAt,
		instance.LastSeen,
	)

	if err != nil {
		zap.L().Error("Failed to update instance", zap.Error(err))
		return err
	}

	return nil
}

// Delete deletes an instance
func (r *InstanceRepository) Delete(id string) error {
	query := `DELETE FROM instances WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		zap.L().Error("Failed to delete instance", zap.Error(err))
		return err
	}

	return nil
}

// UpdateStatus updates the status of an instance
func (r *InstanceRepository) UpdateStatus(id, status string) error {
	query := `UPDATE instances SET status = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.Exec(query, id, status, time.Now())
	if err != nil {
		zap.L().Error("Failed to update instance status", zap.Error(err))
		return err
	}

	return nil
}

// UpdateLastSeen updates the last seen timestamp of an instance
func (r *InstanceRepository) UpdateLastSeen(id string) error {
	query := `UPDATE instances SET last_seen = $2 WHERE id = $1`

	_, err := r.db.Exec(query, id, time.Now())
	if err != nil {
		zap.L().Error("Failed to update instance last seen", zap.Error(err))
		return err
	}

	return nil
}
