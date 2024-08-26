package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/errs"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/beevik/guid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"time"
)

type Store struct {
	conn *sql.DB
}

func (s *Store) Close() error {
	return s.conn.Close()
}

func NewStore(ctx context.Context, conn *sql.DB) (*Store, error) {
	store := &Store{
		conn: conn,
	}

	if err := store.Bootstrap(ctx); err != nil {
		return store, err
	}

	return store, nil
}

func (s *Store) Bootstrap(ctx context.Context) error {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	tx.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS users (
			id varchar(36) PRIMARY KEY,
			user_name varchar(255),
			hash_password varchar(255),
    		deleted boolean DEFAULT FALSE,
			registration_date timestamp
		)`,
	)
	tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS user_name_users_idx ON users (user_name)`)

	tx.ExecContext(
		ctx,
		`CREATE TABLE IF NOT EXISTS datas (
			id varchar(36) PRIMARY KEY,
    		name bytea DEFAULT NULL,
    		user_id varchar(36) NOT NULL,
    		type varchar(100) NOT NULL,
			date timestamp,
			body bytea DEFAULT NULL,
			deleted boolean DEFAULT FALSE,
			description bytea DEFAULT NULL
		)`,
	)
	tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS del_user_type_idx ON datas (deleted, user_id, type)`)

	return tx.Commit()
}

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

func (s *Store) Save(ctx context.Context, data models.Data) (models.Data, error) {
	_, ok, err := s.Get(ctx, data)
	if err != nil {
		return data, fmt.Errorf("ошибка при поиске существующей записи - %w", err)
	}

	if !ok {
		data.ID = guid.NewString()
		data.Date = time.Now()
	}

	fmt.Println("Name", string(data.Name), data.Name)

	_, err = s.conn.ExecContext(ctx, `
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
