package client

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/go-resty/resty/v2"
	"log"
	"time"
)

type RestyClient struct {
	client *resty.Client
	userID string
}

type APIServiceResult struct {
	Code     int
	Response []byte
	Error    error
	Token    string
}

func NewRestyClient() (*RestyClient, error) {
	restyClient := &RestyClient{
		client: resty.New(),
	}

	return restyClient, nil
}

type APIError struct {
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func (c *RestyClient) SetUserID(userID string) {
	c.userID = userID
}

func (c *RestyClient) Send(ctx context.Context, url string, headers map[string]string, data []byte, method string) APIServiceResult {
	result := APIServiceResult{}
	log.Println("source data: ", string(data))
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

	switch method {
	case "PUT":
		response, err = request.Put(url)
	default:
		response, err = request.Post(url)
	}

	if err != nil {
		result.Error = err
	}

	result.Response = response.Body()
	result.Code = response.StatusCode()
	result.Token = response.Header().Get("Authorization")

	return result
}

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

func (c *RestyClient) Close() error {
	return nil
}
