package client

import (
	"context"
)

// Client интерфейс http-клиента.
type Client interface {
	Send(ctx context.Context, url string, headers map[string]string, data []byte, method string) APIServiceResult
	Close() error
	SetUserID(id string)
}
