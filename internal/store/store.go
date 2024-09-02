package store

import (
	"context"
	"github.com/KirillKhitev/goph_keeper/internal/auth"
	"github.com/KirillKhitev/goph_keeper/internal/models"
)

// Интерфейс хранилища данных.
type Store interface {
	CreateUser(ctx context.Context, data auth.AuthorizingData) (models.User, error)
	GetUserByUserName(ctx context.Context, userName string) (models.User, error)
	GetUserByID(ctx context.Context, userID string) (models.User, error)
	Close() error
	Save(ctx context.Context, data models.Data) (models.Data, error)
	Get(ctx context.Context, data models.Data) (models.Data, bool, error)
	List(ctx context.Context, userID string) ([]models.Data, error)
}
