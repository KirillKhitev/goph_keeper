package pg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"reflect"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	type args struct {
		ctx  context.Context
		conn *sql.DB
	}
	tests := []struct {
		name    string
		args    args
		want    *Store
		wantErr bool
	}{
		{
			name: "TestNewStore",
			args: args{
				conn: nil,
			},
			want: &Store{
				conn: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStore(tt.args.ctx, tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStore() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_Close(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	defer db.Close()

	type fields struct {
		conn *sql.DB
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "TestClose",
			fields: fields{
				conn: sqlxDB.DB,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				conn: tt.fields.conn,
			}

			mock.ExpectClose()

			if err := s.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore_CreateUser(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()

	var pgErr pgconn.PgError
	pgErr.Code = pgerrcode.UniqueViolation
	pgErr.Message = "already exist"

	type fields struct {
		conn *sql.DB
	}
	type args struct {
		ctx  context.Context
		data auth.AuthorizingData
		err  pgconn.PgError
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.User
		wantErr bool
	}{
		{
			name: "negative test #1",
			fields: fields{
				conn: sqlxDB.DB,
			},
			args: args{
				ctx: context.Background(),
				data: auth.AuthorizingData{
					UserName: "user3",
					Password: "password",
				},
				err: pgErr,
			},
			want: models.User{
				UserName: "user3",
				ID:       "3",
			},
			wantErr: true,
		},
		{
			name: "positive test #2",
			fields: fields{
				conn: sqlxDB.DB,
			},
			args: args{
				ctx: context.Background(),
				data: auth.AuthorizingData{
					UserName: "user3",
					Password: "password",
				},
			},
			want: models.User{
				UserName: "user3",
				ID:       "3",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				conn: tt.fields.conn,
			}

			var lastInsertID, affected int64
			result := sqlmock.NewResult(lastInsertID, affected)
			ex := mock.ExpectExec("^INSERT (.+)").WillReturnResult(result)

			if tt.args.err.Code != "" {
				ex.WillReturnError(&pgErr)
			}

			fmt.Println(mock.ExpectationsWereMet())

			got, err := s.CreateUser(tt.args.ctx, tt.args.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.UserName, tt.want.UserName) {
				t.Errorf("CreateUser() got = %s, want %s", got.UserName, tt.want.UserName)
			}
		})
	}
}

func TestStore_Get(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()

	type fields struct {
		conn *sql.DB
	}
	type args struct {
		ctx  context.Context
		data models.Data
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.Data
		want1   bool
		wantErr bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				conn: sqlxDB.DB,
			},
			args: args{
				ctx: context.Background(),
				data: models.Data{
					ID:   "111",
					Name: []byte("name"),
					Type: "text",
				},
			},
			want: models.Data{
				ID: "111",
			},
			want1:   true,
			wantErr: false,
		},
		{
			name: "negative test #2",
			fields: fields{
				conn: sqlxDB.DB,
			},
			args: args{
				ctx: context.Background(),
				data: models.Data{
					Name: []byte("name"),
					Type: "text",
				},
			},
			want1:   false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				conn: tt.fields.conn,
			}

			mock.ExpectQuery("^SELECT (.+)").
				WithArgs(tt.args.data.ID).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "user_id", "name", "type", "date", "deleted", "body", "description"}).
						AddRow("111", "user_id", []byte("name"), "text", time.Now(), false, []byte("body"), []byte("description")))

			fmt.Println(mock.ExpectationsWereMet())

			got, got1, err := s.Get(tt.args.ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.ID != tt.want.ID {
				t.Errorf("Get() got = %s, want %s", got.ID, tt.want.ID)
			}

			if got1 != tt.want1 {
				t.Errorf("Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestStore_GetUserByID(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()

	type fields struct {
		conn *sql.DB
	}
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.User
		wantErr bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				conn: sqlxDB.DB,
			},
			args: args{
				ctx:    context.Background(),
				userID: "111",
			},
			want: models.User{
				ID: "111",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				conn: tt.fields.conn,
			}

			mock.ExpectQuery("^SELECT (.+)").
				WithArgs(tt.args.userID).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "user_name", "hash_password", "deleted", "registration_date"}).
						AddRow("111", "login", "sdfsferr4e5345", false, time.Now()))

			fmt.Println(mock.ExpectationsWereMet())

			got, err := s.GetUserByID(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ID != tt.want.ID {
				t.Errorf("GetUserByID() got = %s, want %s", got.ID, tt.want.ID)
			}
		})
	}
}

func TestStore_GetUserByUserName(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()

	type fields struct {
		conn *sql.DB
	}
	type args struct {
		ctx      context.Context
		userName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.User
		wantErr bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				conn: sqlxDB.DB,
			},
			args: args{
				ctx:      context.Background(),
				userName: "login",
			},
			want: models.User{
				UserName: "login",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				conn: tt.fields.conn,
			}

			mock.ExpectQuery("^SELECT (.+)").
				WithArgs(tt.args.userName).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "user_name", "hash_password", "deleted", "registration_date"}).
						AddRow("111", "login", "sdfsferr4e5345", false, time.Now()))

			fmt.Println(mock.ExpectationsWereMet())

			got, err := s.GetUserByUserName(tt.args.ctx, tt.args.userName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByUserName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.UserName != tt.want.UserName {
				t.Errorf("GetUserByUserName() got = %s, want %s", got.UserName, tt.want.UserName)
			}
		})
	}
}

func TestStore_List(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()

	type fields struct {
		conn *sql.DB
	}
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []models.Data
		wantErr bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				conn: sqlxDB.DB,
			},
			args: args{
				ctx:    context.Background(),
				userID: "111",
			},
			want: []models.Data{
				{
					ID: "111",
				},
				{
					ID: "222",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				conn: tt.fields.conn,
			}

			mock.ExpectQuery("^SELECT (.+)").
				WithArgs(tt.args.userID).
				WillReturnRows(
					sqlmock.NewRows([]string{"id", "name", "type", "description"}).
						AddRow("111", []byte("запись1"), "text", []byte("Описание")).
						AddRow("222", []byte("запись2"), "text", []byte("Описание2")))

			fmt.Println(mock.ExpectationsWereMet())

			got, err := s.List(tt.args.ctx, tt.args.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("List() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_createRecord(t *testing.T) {
	db, mock, _ := sqlmock.New()
	sqlxDB := sqlx.NewDb(db, "sqlmock")
	defer db.Close()

	type fields struct {
		conn *sql.DB
	}
	type args struct {
		ctx  context.Context
		data models.Data
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.Data
		wantErr bool
	}{
		{
			name: "positive test #1",
			fields: fields{
				conn: sqlxDB.DB,
			},
			args: args{
				ctx: context.Background(),
				data: models.Data{
					ID:   "111",
					Type: "text",
					Name: []byte("name"),
				},
			},
			want: models.Data{
				ID: "111",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				conn: tt.fields.conn,
			}

			var lastInsertID, affected int64
			result := sqlmock.NewResult(lastInsertID, affected)
			mock.ExpectExec("^INSERT (.+)").WillReturnResult(result)

			fmt.Println(mock.ExpectationsWereMet())

			got, err := s.createRecord(tt.args.ctx, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ID != tt.want.ID {
				t.Errorf("Save() got = %v, want %v", got, tt.want)
			}
		})
	}
}
