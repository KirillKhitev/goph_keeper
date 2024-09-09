package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/go-resty/resty/v2"
	"log"
	"time"
)

// RestyClient структура resty-клиента.
type RestyClient struct {
	client *resty.Client
	userID string
}

// APIServiceResult структура ответа от http-клиента.
type APIServiceResult struct {
	Code     int
	Response []byte
	Error    error
	Token    string
}

// NewRestyClient конструктор resty-клиента.
func NewRestyClient() (*RestyClient, error) {
	restyClient := &RestyClient{
		client: resty.New(),
	}

	return restyClient, nil
}

// APIError структура ошибки от сервера.
type APIError struct {
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// SetUserID сеттер userID в resty-клиент.
func (c *RestyClient) SetUserID(userID string) {
	c.userID = userID
}

func (c *RestyClient) Get(ctx context.Context, headers map[string]string, data []byte) APIServiceResult {
	result := APIServiceResult{}
	log.Println("source data: ", string(data))
	var err error

	url := fmt.Sprintf("http://%s/api/data/get", config.ConfigClient.AddrServer)

	dataForSend, err := c.prepareDataForSend(data)
	if err != nil {
		return result
	}

	ctxt, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	request := c.client.R().
		SetContext(ctxt).
		SetBody(dataForSend).
		SetError(&result.Error).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	for name, value := range headers {
		request.SetHeader(name, value)
	}

	var response *resty.Response

	response, err = request.Post(url)

	if err != nil {
		result.Error = err
	}

	result.Response = response.Body()
	result.Code = response.StatusCode()
	result.Token = response.Header().Get("Authorization")

	return result
}

func (c *RestyClient) Update(ctx context.Context, headers map[string]string, data []byte) APIServiceResult {
	result := APIServiceResult{}
	log.Println("source data: ", string(data))

	url := fmt.Sprintf("http://%s/api/data/update", config.ConfigClient.AddrServer)

	var err error
	dataForSend, err := c.prepareDataForSend(data)
	if err != nil {
		return result
	}

	ctxt, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	request := c.client.R().
		SetContext(ctxt).
		SetBody(dataForSend).
		SetError(&result.Error).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	for name, value := range headers {
		request.SetHeader(name, value)
	}

	var response *resty.Response

	response, err = request.Put(url)

	if err != nil {
		result.Error = err
	}

	result.Response = response.Body()
	result.Code = response.StatusCode()
	result.Token = response.Header().Get("Authorization")

	return result
}

func (c *RestyClient) List(ctx context.Context, headers map[string]string) APIServiceResult {
	result := APIServiceResult{}

	url := fmt.Sprintf("http://%s/api/data/list", config.ConfigClient.AddrServer)

	dataForSend, err := c.prepareDataForSend([]byte{})

	ctxt, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	request := c.client.R().
		SetContext(ctxt).
		SetBody(dataForSend).
		SetError(&result.Error).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	for name, value := range headers {
		request.SetHeader(name, value)
	}

	response, err := request.Post(url)

	if err != nil {
		result.Error = err
	}

	result.Response = response.Body()
	result.Code = response.StatusCode()
	result.Token = response.Header().Get("Authorization")

	return result
}

func (c *RestyClient) Login(ctx context.Context, data []byte) APIServiceResult {
	result := APIServiceResult{}
	log.Println("source data: ", string(data))

	url := fmt.Sprintf("http://%s/api/user/login", config.ConfigClient.AddrServer)

	var err error
	dataForSend, err := c.prepareDataForSend(data)
	if err != nil {
		return result
	}

	ctxt, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	request := c.client.R().
		SetContext(ctxt).
		SetBody(dataForSend).
		SetError(&result.Error).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	var response *resty.Response
	response, err = request.Post(url)

	if err != nil {
		result.Error = err
	}

	result.Response = response.Body()
	result.Code = response.StatusCode()
	result.Token = response.Header().Get("Authorization")

	return result
}

func (c *RestyClient) Register(ctx context.Context, data []byte) APIServiceResult {
	result := APIServiceResult{}
	log.Println("source data: ", string(data))

	url := fmt.Sprintf("http://%s/api/user/register", config.ConfigClient.AddrServer)

	var err error
	dataForSend, err := c.prepareDataForSend(data)
	if err != nil {
		return result
	}

	ctxt, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	request := c.client.R().
		SetContext(ctxt).
		SetBody(dataForSend).
		SetError(&result.Error).
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip")

	var response *resty.Response

	response, err = request.Post(url)

	if err != nil {
		result.Error = err
	}

	result.Response = response.Body()
	result.Code = response.StatusCode()
	result.Token = response.Header().Get("Authorization")

	return result
}

// prepareDataForSend подготавливает данные перед отправкой на сервер.
func (c *RestyClient) prepareDataForSend(data []byte) ([]byte, error) {
	dataCompress, err := c.Compress(data)
	if err != nil {
		return dataCompress, fmt.Errorf("error by compress data: %s, err: %w", data, err)
	}

	return dataCompress, nil
}

// Compress сжимает данные перед отправкой на сервер.
func (c *RestyClient) Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer

	w := gzip.NewWriter(&b)

	_, err := w.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}

	err = w.Close()
	if err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}

	return b.Bytes(), nil
}

// Close закрытие resty-клиента.
func (c *RestyClient) Close() error {
	return nil
}
