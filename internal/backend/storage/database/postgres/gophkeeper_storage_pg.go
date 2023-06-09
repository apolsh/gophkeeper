package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/apolsh/yapr-gophkeeper/internal/backend/service"
	"github.com/apolsh/yapr-gophkeeper/internal/misc/db"
	"github.com/apolsh/yapr-gophkeeper/internal/model"
	errs "github.com/apolsh/yapr-gophkeeper/internal/model/app_errors"
	"github.com/apolsh/yapr-gophkeeper/internal/model/dto"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var _ service.SecretStorage = (*GophkeeperStoragePG)(nil)
var _ service.UserStorage = (*GophkeeperStoragePG)(nil)

const (
	constraintUniqUsername = "clients_username_key"
)

//go:embed migrations/*.sql
var fs embed.FS

// GophkeeperStoragePG user and order storage postgres implementation.
type GophkeeperStoragePG struct {
	db *pgxpool.Pool
}

// NewGophkeeperStoragePG GophkeeperStoragePG constructor.
func NewGophkeeperStoragePG(databaseDSN string) (*GophkeeperStoragePG, error) {
	conn, err := pgxpool.Connect(context.Background(), databaseDSN)
	if err != nil {
		return nil, fmt.Errorf(`repository initialization error: %w`, err)
	}
	migrationSourceDriver, err := iofs.New(fs, "migrations")
	if err != nil {
		return nil, fmt.Errorf(`failed to iniatilise migration source: %w`, err)
	}

	err = db.RunMigration(migrationSourceDriver, databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return &GophkeeperStoragePG{
		db: conn,
	}, nil
}

// NewUser saves new user.
func (s *GophkeeperStoragePG) NewUser(ctx context.Context, login string, hashedPassword string) (model.User, error) {
	var id int64
	timestamp := time.Now().UTC().UnixMilli()
	q := "INSERT INTO clients (username, password, date_last_modified) VALUES ($1, $2, $3)  RETURNING client_id"
	err := s.db.QueryRow(ctx, q, login, hashedPassword, timestamp).Scan(&id)

	var pgErr *pgconn.PgError
	if err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.ConstraintName == constraintUniqUsername {
				return model.User{}, errs.ErrorLoginIsAlreadyUsed
			}
		}
		return model.User{}, errs.HandleUnknownDatabaseError(err)
	}

	return model.User{ID: id, Login: login, HashedPassword: hashedPassword, Timestamp: timestamp}, nil
}

// GetUserByLogin returns user by login.
func (s *GophkeeperStoragePG) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	q := "SELECT client_id, username, password FROM clients WHERE username = $1"
	var user model.User
	err := s.db.QueryRow(ctx, q, login).Scan(&user.ID, &user.Login, &user.HashedPassword)
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return model.User{}, errs.ErrItemNotFound
		}
		return model.User{}, errs.HandleUnknownDatabaseError(err)
	}
	return user, nil
}

// GetSecretSyncMetaByUser returns metadata for secret synchronization.
func (s *GophkeeperStoragePG) GetSecretSyncMetaByUser(ctx context.Context, userID int64) ([]dto.SecretSyncMetadata, error) {
	q := "SELECT secret_id, hash, date_last_modified FROM secrets WHERE owner = $1"

	rows, err := s.db.Query(ctx, q, userID)
	if err != nil {
		return nil, errs.HandleUnknownDatabaseError(err)
	}

	secretSyncMetas := make([]dto.SecretSyncMetadata, 0)
	var secretSyncMeta dto.SecretSyncMetadata
	for rows.Next() {
		err := rows.Scan(&secretSyncMeta.ID, &secretSyncMeta.Hash, &secretSyncMeta.Timestamp)
		if err != nil {
			return nil, errs.HandleUnknownDatabaseError(err)
		}
		secretSyncMetas = append(secretSyncMetas, secretSyncMeta)
	}

	return secretSyncMetas, nil
}

func (s *GophkeeperStoragePG) GetSecretSyncMetaByOwnerAndName(ctx context.Context, userID int, name string) (dto.SecretSyncMetadata, error) {
	q := "SELECT secret_id, hash, date_last_modified FROM secrets WHERE owner = $1 AND name = $2"

	var secretSyncMeta dto.SecretSyncMetadata
	err := s.db.QueryRow(ctx, q, userID, name).Scan(&secretSyncMeta.ID, &secretSyncMeta.Hash, &secretSyncMeta.Timestamp)
	if err != nil {
		return secretSyncMeta, errs.HandleUnknownDatabaseError(err)
	}
	return secretSyncMeta, nil
}

// GetSecretByID returns EncodedSecret by ID.
func (s *GophkeeperStoragePG) GetSecretByID(ctx context.Context, userID int, secretID string) (model.EncodedSecret, error) {
	q := "SELECT secret_id, owner, name, hash, description, enc_data, type, date_last_modified FROM secrets WHERE secret_id = $1 AND owner = $2"

	var encSecret model.EncodedSecret

	err := s.db.QueryRow(ctx, q, secretID, userID).Scan(
		&encSecret.ID,
		&encSecret.Owner,
		&encSecret.Name,
		&encSecret.Hash,
		&encSecret.Description,
		&encSecret.EncodedContent,
		&encSecret.Type,
		&encSecret.Timestamp)

	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return model.EncodedSecret{}, errs.ErrItemNotFound
		}
		return model.EncodedSecret{}, errs.HandleUnknownDatabaseError(err)
	}
	return encSecret, nil
}

// SaveEncodedSecret saves new EncodedSecret.
func (s *GophkeeperStoragePG) SaveEncodedSecret(ctx context.Context, secret model.EncodedSecret) error {
	q := "INSERT INTO secrets (secret_id, owner, name, hash, description, enc_data, type, date_last_modified) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"

	_, err := s.db.Exec(ctx, q, secret.ID, secret.Owner, secret.Name, secret.Hash, secret.Description, secret.EncodedContent, secret.Type, secret.Timestamp)
	if err != nil {
		return errs.HandleUnknownDatabaseError(err)
	}

	return nil
}

// DeleteEncodedSecret deletes EncodedSecret.
func (s *GophkeeperStoragePG) DeleteEncodedSecret(ctx context.Context, ownerID int, secretID string) error {
	q := "DELETE FROM secrets WHERE secret_id = $1 AND owner = $2"
	_, err := s.db.Exec(ctx, q, secretID, ownerID)
	if err != nil {
		return errs.HandleUnknownDatabaseError(err)
	}

	return nil
}

// Close closes database connection.
func (s *GophkeeperStoragePG) Close() {
	s.db.Close()
}
