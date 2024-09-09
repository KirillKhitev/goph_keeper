package client

import (
	"context"
)

// Client интерфейс http-клиента.
type Client interface {
	Get(ctx context.Context, headers map[string]string, data []byte) APIServiceResult
	Update(ctx context.Context, headers map[string]string, data []byte) APIServiceResult
	List(ctx context.Context, headers map[string]string) APIServiceResult
	Login(ctx context.Context, data []byte) APIServiceResult
	Register(ctx context.Context, data []byte) APIServiceResult
	Close() error
	SetUserID(id string)
}
