package main

import (
	"flag"
	"os"
	"strconv"
)

var (
	// serverAddress указывает адрес сервера метрик.
	serverAddress string

	// pollInterval задаёт интервал опроса в секундах.
	pollInterval int

	// reportInterval задаёт интервал отправки отчётов в секундах.
	reportInterval int

	// key используется для генерации HMAC SHA256 хеша.
	key string

	// rateLimit ограничивает максимальное количество одновременных исходящих запросов.
	rateLimit int

	// batchSize задаёт размер батча для отправки данных.
	batchSize int

	// logLevel задаёт уровень логирования.
	logLevel string

	// header содержит имя HTTP-заголовка для передачи хеша.
	header string
)

// parseFlags считывает параметры из командной строки и переменных окружения,
// инициализируя конфигурацию клиента.
func parseFlags() {
	flag.StringVar(&serverAddress, "a", "http://localhost:8080", "Metrics server address")
	flag.IntVar(&pollInterval, "p", 2, "Poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "Report interval in seconds")
	flag.StringVar(&key, "k", "", "Key for HMAC SHA256 hash")
	flag.IntVar(&rateLimit, "l", 0, "Max number of concurrent outgoing requests")

	flag.Parse()

	// Переопределение значений переменными окружения, если они заданы.
	if env := os.Getenv("ADDRESS"); env != "" {
		serverAddress = env
	}
	if env := os.Getenv("POLL_INTERVAL"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			pollInterval = v
		}
	}
	if env := os.Getenv("REPORT_INTERVAL"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			reportInterval = v
		}
	}
	if env := os.Getenv("KEY"); env != "" {
		key = env
	}
	if env := os.Getenv("RATE_LIMIT"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			rateLimit = v
		}
	}

	// Значения по умолчанию.
	logLevel = "info"
	header = "HashSHA256"
	batchSize = 100
}
