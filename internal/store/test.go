package store

import (
	"context"
	"errors"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/errs"
	"github.com/KirillKhitev/goph_keeper/internal/models"
	"time"
)

type TestStore struct{}

func GetTestStore() *TestStore {
	ts := &TestStore{}

	return ts
}
func (t *TestStore) Get(ctx context.Context, data models.Data) (models.Data, bool, error) {
	result := models.Data{
		ID:   "11122333",
		Name: []byte("Hello World"),
	}

	if data.ID != result.ID {
		return models.Data{}, false, nil
	}

	return result, true, nil
}

func (t *TestStore) CreateUser(ctx context.Context, data auth.AuthorizingData) (models.User, error) {
	if data.UserName == "Exist User" {
		return models.User{}, errs.ErrAlreadyExist
	}

	result := models.User{
		ID:               "111",
		UserName:         data.UserName,
		HashPassword:     data.GenerateHashPassword(),
		RegistrationDate: time.Now(),
	}

	return result, nil
}

func (t *TestStore) GetUserByUserName(ctx context.Context, userName string) (models.User, error) {
	if userName == "" {
		return models.User{}, errors.New("userName is empty")
	}

	if userName != "Exist User" {
		return models.User{}, errors.New("user not found")
	}

	result := models.User{
		ID:               "111",
		UserName:         userName,
		HashPassword:     "e7f900a989cc919bca22bb0a7df18113d2930f8f9780db1233b2afa5a97ce7f7",
		RegistrationDate: time.Now(),
	}

	return result, nil
}

func (t *TestStore) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	if userID == "" {
		return models.User{}, errors.New("invalid user id")
	}

	if userID != "111" {
		return models.User{}, errors.New("user not found")
	}

	result := models.User{
		ID:               "111",
		UserName:         "Пользователь1",
		RegistrationDate: time.Now(),
	}

	return result, nil
}
func (t *TestStore) Close() error {
	return nil
}

func (t *TestStore) Save(ctx context.Context, data models.Data) (models.Data, error) {
	if data.UserID == "" {
		return models.Data{}, errors.New("empty userID")
	}

	return data, nil
}

func (t *TestStore) List(ctx context.Context, userID string) ([]models.Data, error) {
	if userID == "" {
		return []models.Data{}, errors.New("пустой userID")
	}

	result := []models.Data{
		{
			ID:     "2342343",
			Name:   []byte("Hello World"),
			Type:   "credit_card",
			UserID: userID,
		},
	}

	return result, nil
}
