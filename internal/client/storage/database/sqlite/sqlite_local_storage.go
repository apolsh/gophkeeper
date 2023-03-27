package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/apolsh/yapr-gophkeeper/internal/client/controller"
	"github.com/apolsh/yapr-gophkeeper/internal/client/storage"
	"github.com/apolsh/yapr-gophkeeper/internal/logger"
	"github.com/apolsh/yapr-gophkeeper/internal/misc/db"
	"github.com/apolsh/yapr-gophkeeper/internal/model"
	"github.com/apolsh/yapr-gophkeeper/internal/model/dto"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/mattn/go-sqlite3"
)

const databaseName = "gophkeeper.db"

//go:embed migrations/*.sql
var fs embed.FS

var log = logger.LoggerOfComponent("sqlite-local-storage")

var _ controller.LocalStorage = (*GophkeeperLocalStorageSqlite)(nil)

// GophkeeperLocalStorageSqlite implementation of LocalStorage.
type GophkeeperLocalStorageSqlite struct {
	db *sql.DB
}

// NewGophkeeperLocalStorageSqlite GophkeeperLocalStorageSqlite constructor.
func NewGophkeeperLocalStorageSqlite(applicationDir string) (*GophkeeperLocalStorageSqlite, error) {
	databasePath := filepath.Join(applicationDir, databaseName)

	database, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf(`repository initialization error: %w`, err)
	}

	migrationSourceDriver, err := iofs.New(fs, "migrations")
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf(`failed to iniatilise migration source: %w`, err)
	}

	err = db.RunMigration(migrationSourceDriver, fmt.Sprintf("sqlite3://%s", databasePath))
	if err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return &GophkeeperLocalStorageSqlite{
		db: database,
	}, nil
}

// SaveUser saves new user.
func (g GophkeeperLocalStorageSqlite) SaveUser(ctx context.Context, user model.User) error {
	q := "INSERT INTO clients (client_id, username, password, date_last_modified) VALUES ($1, $2, $3, $4)"
	_, err := g.db.ExecContext(ctx, q, user.ID, user.Login, user.HashedPassword, user.Timestamp)
	if err != nil {
		return err
	}
	return nil
}

// UpdateUser update users metadata.
func (g GophkeeperLocalStorageSqlite) UpdateUser(ctx context.Context, user model.User) error {
	q := "UPDATE clients SET username = $1, password = $2, date_last_modified = $3 WHERE client_id = $4"
	_, err := g.db.ExecContext(ctx, q, user.Login, user.HashedPassword, user.Timestamp, user.ID)
	if err != nil {
		return err
	}
	return nil
}

// GetUserByID returns user by ID.
func (g GophkeeperLocalStorageSqlite) GetUserByID(ctx context.Context, userID int64) (user model.User, err error) {
	q := "SELECT client_id, username, password, date_last_modified FROM clients WHERE client_id = $1"
	row := g.db.QueryRowContext(ctx, q, userID)
	if err != nil {
		return model.User{}, err
	}
	err = row.Scan(&user.ID, &user.Login, &user.HashedPassword, &user.Timestamp)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return model.User{}, storage.ErrorItemNotFound
		}
	}
	return
}

// GetSecretSyncMetaByID returns metadata for one secret synchronization.
func (g GophkeeperLocalStorageSqlite) GetSecretSyncMetaByID(ctx context.Context, id string) (secretMeta dto.SecretSyncMetadata, err error) {
	q := "SELECT secret_id, hash, date_last_modified FROM secrets WHERE secret_id = $1"
	row := g.db.QueryRowContext(ctx, q, id)
	err = row.Scan(&secretMeta.ID, &secretMeta.Hash, &secretMeta.Timestamp)
	return
}

// GetSecretSyncMetaByOwnerID returns metadata for all secrets synchronization by user.
func (g GophkeeperLocalStorageSqlite) GetSecretSyncMetaByOwnerID(ctx context.Context, ownerID int64) ([]dto.SecretSyncMetadata, error) {
	secretSyncMetas := make([]dto.SecretSyncMetadata, 0, 0)

	q := "SELECT secret_id, hash, date_last_modified FROM secrets WHERE owner = $1"
	rows, err := g.db.QueryContext(ctx, q, ownerID)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Error(err)
		}
	}(rows)

	for rows.Next() {
		var secretSyncMeta dto.SecretSyncMetadata
		err := rows.Scan(&secretSyncMeta.ID, &secretSyncMeta.Hash, &secretSyncMeta.Timestamp)
		if err != nil {
			return nil, err
		}
		secretSyncMetas = append(secretSyncMetas, secretSyncMeta)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return secretSyncMetas, nil
}

// SaveEncodedSecret saves EncodedSecret.
func (g GophkeeperLocalStorageSqlite) SaveEncodedSecret(ctx context.Context, encSecret model.EncodedSecret) error {
	q := "INSERT INTO secrets (secret_id, owner, name, hash, description, enc_data, type, date_last_modified) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"
	_, err := g.db.ExecContext(ctx, q, encSecret.ID, encSecret.Owner, encSecret.Name, encSecret.Hash, encSecret.Description, encSecret.EncodedContent, encSecret.Type, encSecret.Timestamp)
	if err != nil {
		return err
	}
	return nil
}

// GetSecretByID returns EncodedSecret by ID.
func (g GophkeeperLocalStorageSqlite) GetSecretByID(ctx context.Context, id string) (encSecret model.EncodedSecret, err error) {
	q := "SELECT secret_id, owner, name, hash, description, enc_data, type, date_last_modified FROM secrets WHERE secret_id = $1"
	row := g.db.QueryRowContext(ctx, q, id)
	err = row.Scan(
		&encSecret.ID,
		&encSecret.Owner,
		&encSecret.Name,
		&encSecret.Hash,
		&encSecret.Description,
		&encSecret.EncodedContent,
		&encSecret.Type,
		&encSecret.Timestamp)
	return
}

// GetAllSecretsItemInfoByUserID returns all secret item info by user id.
func (g GophkeeperLocalStorageSqlite) GetAllSecretsItemInfoByUserID(ctx context.Context, ownerID int64) ([]dto.SecretItemInfo, error) {
	secretInfos := make([]dto.SecretItemInfo, 0, 0)

	q := "SELECT name, description, type FROM secrets WHERE owner = $1"
	rows, err := g.db.QueryContext(ctx, q, ownerID)
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Error(err)
		}
	}(rows)
	for rows.Next() {
		var secretInfo dto.SecretItemInfo
		err := rows.Scan(&secretInfo.Name, &secretInfo.Description, &secretInfo.SecretType)
		if err != nil {
			return nil, err
		}
		secretInfos = append(secretInfos, secretInfo)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return secretInfos, nil
}

// GetSecretByName returns secret by it name.
func (g GophkeeperLocalStorageSqlite) GetSecretByName(ctx context.Context, name string) (encSecret model.EncodedSecret, err error) {
	q := "SELECT secret_id, owner, name, hash, description, enc_data, type, date_last_modified FROM secrets WHERE name = $1"
	row := g.db.QueryRowContext(ctx, q, name)
	err = row.Scan(
		&encSecret.ID,
		&encSecret.Owner,
		&encSecret.Name,
		&encSecret.Hash,
		&encSecret.Description,
		&encSecret.EncodedContent,
		&encSecret.Type,
		&encSecret.Timestamp)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return model.EncodedSecret{}, storage.ErrorItemNotFound
		}
	}
	return
}

// DeleteSecretByName delete secret by it name.
func (g GophkeeperLocalStorageSqlite) DeleteSecretByName(ctx context.Context, name string) (id string, err error) {
	tx, err := g.db.BeginTx(ctx, nil)
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			log.Error(err)
		}
	}(tx)
	idQuery := "SELECT secret_id FROM secrets WHERE name = $1"
	row := tx.QueryRowContext(ctx, idQuery, name)
	err = row.Scan(&id)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return "", storage.ErrorItemNotFound
		}
		return
	}
	deleteQuery := "DELETE FROM secrets WHERE name = $1"
	_, err = tx.ExecContext(ctx, deleteQuery, name)
	if err != nil {
		return
	}
	err = tx.Commit()
	return
}
