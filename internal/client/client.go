package client

import (
	"context"
)

// Client интерфейс http-клиента.
type Client interface {
	Get(ctx context.Context, url string, headers map[string]string, data []byte) APIServiceResult
	Update(ctx context.Context, url string, headers map[string]string, data []byte) APIServiceResult
	List(ctx context.Context, url string, headers map[string]string) APIServiceResult
	Login(ctx context.Context, url string, data []byte) APIServiceResult
	Register(ctx context.Context, url string, data []byte) APIServiceResult
	Close() error
	SetUserID(id string)
}
