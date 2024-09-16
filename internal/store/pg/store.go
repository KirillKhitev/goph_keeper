// Пакет Postgre-хранилища данных.
package pg

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/KirillKhitev/goph_keeper/internal/errs"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/beevik/guid"
	"github.com/go-pg/migrations/v8"
	"github.com/go-pg/pg/v10"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
	"time"
)

// Структура хранилища.
type Store struct {
	conn *sql.DB
}

// Close закрывает хранилище.
func (s *Store) Close() error {
	return s.conn.Close()
}

//go:embed migrations/*.sql
var migs embed.FS

const directory_migrations = "migrations"

// Конструктор хранилища.
func NewStore(ctx context.Context, conn *sql.DB) (*Store, error) {
	store := &Store{
		conn: conn,
	}

	opt := &pg.Options{
		User:     config.ConfigServer.MigrationUser,
		Password: config.ConfigServer.MigrationPassword,
		Database: config.ConfigServer.MigrationDB,
	}

	db := pg.Connect(opt)
	collection := migrations.NewCollection()
	collection.DiscoverSQLMigrationsFromFilesystem(http.FS(migs), "migrations")
	collection.DisableSQLAutodiscover(true)
	collection.Run(db, "init", "version")
	collection.Run(db)

	return store, nil
}

// CreateUser создает нового пользователя.
func (s *Store) CreateUser(ctx context.Context, data auth.AuthorizingData) (models.User, error) {
	user := data.NewUserFromData()

	_, err := s.conn.ExecContext(
		ctx,
		`INSERT INTO users
			(id, user_name, hash_password, registration_date)
		VALUES
			($1, $2, $3, $4)`,
		user.ID, data.UserName, user.HashPassword, user.RegistrationDate,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return user, errs.ErrAlreadyExist
		}
	}

	return user, err
}

// GetUserByUserName ищет пользователя в БД по user_name.
func (s *Store) GetUserByUserName(ctx context.Context, userName string) (models.User, error) {
	var user models.User

	row := s.conn.QueryRowContext(
		ctx,
		`SELECT
			*
		FROM
			users
		WHERE
		    deleted = false AND
			user_name = $1`,
		userName,
	)

	err := row.Scan(&user.ID, &user.UserName, &user.HashPassword, &user.Deleted, &user.RegistrationDate)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return user, errs.ErrNotFound
	}

	return user, err
}

// GetUserByID ищет пользователя в БД по ID.
func (s *Store) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	var user models.User

	row := s.conn.QueryRowContext(
		ctx,
		`SELECT
			*
		FROM
			users
		WHERE
			id = $1`,
		userID,
	)

	err := row.Scan(&user.ID, &user.UserName, &user.HashPassword, &user.Deleted, &user.RegistrationDate)
	if err != nil {
		return user, err
	}

	return user, nil
}

// List получает список записей пользователя.
func (s *Store) List(ctx context.Context, userID string) ([]models.Data, error) {
	var result []models.Data

	rows, err := s.conn.QueryContext(
		ctx,
		`SELECT
			id,
			name,
    		type,
			description
		FROM
			datas
		WHERE
			deleted = false AND
			user_id = $1`,
		userID,
	)

	if err != nil {
		return result, fmt.Errorf("unable query: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		item := models.Data{}

		err = rows.Scan(&item.ID, &item.Name, &item.Type, &item.Description)
		if err != nil {
			return result, fmt.Errorf("unable to scan row: %w", err)
		}

		result = append(result, item)
	}

	if err := rows.Err(); err != nil {
		return result, fmt.Errorf("cursor error: %w", err)
	}

	return result, nil
}

// Get получает запись пользователя.
func (s *Store) Get(ctx context.Context, data models.Data) (models.Data, bool, error) {
	var result models.Data

	if data.ID == "" {
		return result, false, nil
	}

	row := s.conn.QueryRowContext(
		ctx,
		`SELECT
			id,
			user_id,
			name,
			type,
			date,
    		deleted,
    		body,
    		description
		FROM
			datas
		WHERE
			id = $1`,
		data.ID,
	)

	err := row.Scan(&result.ID, &result.UserID, &result.Name, &result.Type, &result.Date, &result.Deleted, &result.Body, &result.Description)
	if err != nil {
		return result, false, err
	}

	return result, true, nil
}

// Save сохраняет запись в БД.
func (s *Store) Save(ctx context.Context, data models.Data) (models.Data, error) {
	_, ok, err := s.Get(ctx, data)
	if err != nil {
		return data, fmt.Errorf("ошибка при поиске существующей записи - %w", err)
	}

	if !ok {
		data.ID = guid.NewString()
		data.Date = time.Now()
	}

	data, err = s.createRecord(ctx, data)

	return data, err
}

func (s *Store) createRecord(ctx context.Context, data models.Data) (models.Data, error) {
	_, err := s.conn.ExecContext(ctx, `
	INSERT INTO datas
	    (id, name, user_id, type, date, body, deleted, description)
	VALUES
	    ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (id) DO UPDATE SET
	    id = $1,
	    name = $2,
	    user_id = $3,
	    type = $4,
	    date = $5,
	    body = $6,
	    deleted = $7,
	    description = $8
	`,
		data.ID,
		data.Name,
		data.UserID,
		data.Type,
		data.Date,
		data.Body,
		data.Deleted,
		data.Description,
	)

	if err != nil {
		return data, fmt.Errorf("unable to save row: %w", err)
	}

	return data, nil
}
