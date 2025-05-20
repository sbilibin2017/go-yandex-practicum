package facades

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/sbilibin2017/go-yandex-practicum/internal/types"
)

// MetricFacade инкапсулирует логику отправки метрик на внешний HTTP-сервер.
// Поддерживает сериализацию, установку заголовков и вычисление хеша тела запроса.
type MetricFacade struct {
	client        *resty.Client
	marshalerFunc func(v any) ([]byte, error)
	key           string
	header        string
	hashFunc      func(data []byte, key string) string
}

// NewMetricFacade создает новый экземпляр MetricFacade.
//
// client — HTTP-клиент (resty), который будет использоваться для отправки запросов.
// marshalerFunc — функция сериализации метрик (например, json.Marshal).
// hashFunc — функция хеширования тела запроса (например, HMAC).
// serverAddress — адрес сервера, на который отправляются метрики (например, "localhost:8080").
// key — секретный ключ для хеширования тела запроса.
// header — имя заголовка, в который будет помещён хеш.
//
// Возвращает готовый к использованию *MetricFacade.
func NewMetricFacade(
	client *resty.Client,
	marshalerFunc func(v any) ([]byte, error),
	hashFunc func(data []byte, key string) string,
	serverAddress string,
	key string,
	header string,
) *MetricFacade {
	if !strings.HasPrefix(serverAddress, "http://") && !strings.HasPrefix(serverAddress, "https://") {
		serverAddress = "http://" + serverAddress
	}
	client = client.SetBaseURL(serverAddress)
	return &MetricFacade{
		client:        client,
		marshalerFunc: marshalerFunc,
		key:           key,
		header:        header,
		hashFunc:      hashFunc,
	}
}

// Updates отправляет срез метрик на сервер по эндпоинту "/updates/".
// ctx — контекст для отмены или таймаута запроса.
// metrics — срез метрик, сериализуемый и передаваемый в теле запроса.
//
// Если метрики не переданы, функция ничего не делает.
// Возвращает ошибку, если не удалось сериализовать метрики,
// отправить запрос или если сервер вернул ошибку.
func (mf *MetricFacade) Updates(
	ctx context.Context,
	metrics []types.Metrics,
) error {
	if len(metrics) == 0 {
		return nil
	}

	bodyBytes, err := mf.marshalerFunc(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	req := mf.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(bodyBytes)

	mf.setHashHeader(req, bodyBytes)

	resp, err := req.Post("/updates/")
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	if resp.IsError() {
		return fmt.Errorf("error response from server for metrics: %s", resp.String())
	}
	return nil
}

// setHashHeader добавляет в HTTP-запрос заголовок с хешем тела запроса.
// Используется только если задан ключ и функция хеширования.
func (mf *MetricFacade) setHashHeader(req *resty.Request, body []byte) {
	if mf.key == "" || mf.hashFunc == nil {
		return
	}
	hashValue := mf.hashFunc(body, mf.key)
	req.SetHeader(mf.header, hashValue)
}
