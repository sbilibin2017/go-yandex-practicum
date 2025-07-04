
# Яндекс Практикум "Продвинутый Go Разработчик"  
## Сервис сбора метрик и алертинга

## 📌 Описание проекта

Это система для сбора runtime-метрик, разработанная в рамках трека **«Сервис сбора метрик и алертинга»** курса **«Продвинутый Go-разработчик»** от Яндекс Практикума.  
Состоит из двух компонентов:

- **Сервер** — принимает и хранит метрики;
- **Агент** — собирает метрики из среды выполнения и отправляет их на сервер.

---

## ⚙️ Метрики

Поддерживаются два типа метрик:

- **gauge** (`float64`) — значение перезаписывается при каждом обновлении;
- **counter** (`int64`) — значение увеличивается на заданное при каждом обновлении.

---

## 🌐 Сервер

Сервер слушает `http://localhost:8080` и обрабатывает http запросы  

---

## 🛰 Агент

Агент предназначен для автоматического сбора метрик из стандартной библиотеки `runtime` и их периодической отправки на сервер.

### Что делает агент:

- Использует `runtime.ReadMemStats` для сбора метрик (`Alloc`, `TotalAlloc`, `Sys`, и др.);
- Обновляет значения `counter` метрик (например, количество попыток отправки);
- Отправляет данные на сервер с заданной периодичностью (`pollInterval`, `reportInterval`);
- Работает параллельно, используя `context.Context` и фоновые воркеры.

Агент можно запустить отдельно, указав адрес сервера и частоту опроса/отправки метрик через конфигурацию.

---

## Структура проекта

Проект имеет слоистую архитектуру, взаимодействие слоев осуществляется с помощью интерфейсов


```
.                                                       # Корень проекта
├── cmd                                                 # Точки входа: отдельные исполняемые компоненты
│   ├── agent                                           # Агент для сбора и отправки метрик
│   │   ├── build_info.go                               # Содержит информацию о сборке агента (версия, время)
│   │   ├── flags.go                                    # Обработка командных флагов агента
│   │   ├── main.go                                     # Точка входа агента
│   │   └── run.go                                      # Логика инициализации и запуска агента
│   ├── server                                          # HTTP-сервер для приёма и хранения метрик
│   │   ├── build_info.go                               # Содержит информацию о сборке сервера
│   │   ├── flags.go                                    # Обработка CLI-флагов сервера
│   │   ├── main.go                                     # Главная функция сервера
│   │   └── run.go                                      # Логика инициализации и запуска сервера
│   └── staticlint                                      # Утилита статического анализа кода
│       ├── analyzers                                   # Кастомные анализаторы
│       │   └── noexit                                  # Анализатор, запрещающий прямой вызов os.Exit
│       │       └── noexit.go                           # Реализация анализатора noexit
│       └── main.go                                     # Запуск утилиты staticlint
├── go.mod                                              # Зависимости и настройки Go-модуля
├── go.sum                                              # Контрольные суммы зависимостей

├── internal                                            # Внутренние пакеты (недоступны вне проекта)
│   ├── contexts                                        # Работа с контекстами и транзакциями
│   │   ├── tx.go                                       # Расширение context с поддержкой транзакций
│   │   └── tx_test.go                                  # Тесты для tx.go
│
│   ├── facades                                         # Фасады (упрощённые интерфейсы к сервисам)
│   │   ├── metric.go                                   # Интерфейс фасада для работы с метриками
│   │   └── metric_test.go                              # Тесты фасада
│
│   ├── handlers                                        # HTTP-обработчики API
│   │   ├── db_ping.go                                  # Обработчик healthcheck'а базы данных
│   │   ├── db_ping_test.go                             # Тесты для db_ping
│   │   ├── helpers.go                                  # Вспомогательные функции
│   │   ├── helpers_test.go                             # Тесты для helpers
│   │   ├── metric_list_all_html.go                     # Возврат всех метрик в HTML-формате
│   │   ├── metric_list_all_html_mock.go                # Моки для HTML-обработчика
│   │   ├── metric_list_all_html_test.go                # Тесты для HTML-обработчика
│   │   ├── metric_update_body.go                       # Обновление одной метрики из JSON в теле
│   │   ├── metric_update_body_mock.go                  # Моки для update_body
│   │   ├── metric_update_body_test.go                  # Тесты update_body
│   │   ├── metric_update_path.go                       # Обновление одной метрики из URL-параметра
│   │   ├── metric_update_path_mock.go                  # Моки
│   │   ├── metric_update_path_test.go                  # Тесты
│   │   ├── metric_updates_body.go                      # Обновление нескольких метрик из тела
│   │   ├── metric_updates_body_mock.go                 # Моки
│   │   ├── metric_updates_body_test.go                 # Тесты
│   │   ├── metric_value_body.go                        # Получение значения метрики из тела
│   │   ├── metric_value_body_mock.go                   # Моки
│   │   ├── metric_value_body_test.go                   # Тесты
│   │   ├── metric_value_path.go                        # Получение метрики по имени/типу в URL
│   │   ├── metric_value_path_mock.go                   # Моки
│   │   └── metric_value_path_test.go                   # Тесты
│
│   ├── logger                                          # Настройка логирования
│   │   ├── logger.go                                   # Инициализация и конфигурация логгера
│   │   └── logger_test.go                              # Тесты логгера
│
│   ├── middlewares                                     # HTTP-мидлвары
│   │   ├── crypto.go                                   # Шифрование и подпись
│   │   ├── crypto_test.go                              # Тесты crypto
│   │   ├── gzip.go                                     # Сжатие ответа
│   │   ├── gzip_test.go                                # Тесты gzip
│   │   ├── hash.go                                     # Проверка целостности
│   │   ├── hash_test.go                                # Тесты hash
│   │   ├── logging.go                                  # Мидлвар логирования запросов
│   │   ├── logging_test.go                             # Тесты logging
│   │   ├── retry.go                                    # Повтор при ошибках
│   │   ├── retry_test.go                               # Тесты retry
│   │   ├── tx.go                                       # Обёртка транзакции
│   │   └── tx_test.go                                  # Тесты tx
│
│   ├── repositories                                    # Слой доступа к данным (БД, память, файлы)
│   │   ├── helpers.go                                  # Вспомогательные функции
│   │   ├── helpers_test.go                             # Тесты helpers
│   │   ├── metric_get_by_id_db.go                      # Получение метрики по ID из БД
│   │   ├── metric_get_by_id_db_test.go                 # Тесты
│   │   ├── metric_get_by_id_file.go                    # Из файла
│   │   ├── metric_get_by_id_file_test.go               # Тесты
│   │   ├── metric_get_by_id.go                         # Общее поведение
│   │   ├── metric_get_by_id_memory.go                  # Из памяти
│   │   ├── metric_get_by_id_memory_test.go             # Тесты
│   │   ├── metric_get_by_id_mock.go                    # Моки
│   │   ├── metric_get_by_id_test.go                    # Общие тесты
│   │   ├── metric_list_all_db.go                       # Получение всех метрик из БД
│   │   ├── metric_list_all_db_test.go                  # Тесты
│   │   ├── metric_list_all_file.go                     # Из файла
│   │   ├── metric_list_all_file_test.go                # Тесты
│   │   ├── metric_list_all.go                          # Общая логика
│   │   ├── metric_list_all_memory.go                   # Из памяти
│   │   ├── metric_list_all_memory_test.go              # Тесты
│   │   ├── metric_list_all_mock.go                     # Моки
│   │   ├── metric_list_all_test.go                     # Общие тесты
│   │   ├── metric_save_db.go                           # Сохранение в БД
│   │   ├── metric_save_db_test.go                      # Тесты
│   │   ├── metric_save_file.go                         # В файл
│   │   ├── metric_save_file_test.go                    # Тесты
│   │   ├── metric_save.go                              # Общая логика
│   │   ├── metric_save_memory.go                       # В память
│   │   ├── metric_save_memory_test.go                  # Тесты
│   │   ├── metric_save_mock.go                         # Моки
│   │   └── metric_save_test.go                         # Тесты
│
│   ├── runners                                         # Управление жизненным циклом компонентов
│   │   ├── server.go                                   # Запуск сервера
│   │   ├── server_mock.go                              # Моки
│   │   ├── server_test.go                              # Тесты
│   │   ├── worker.go                                   # Запуск агента (воркера)
│   │   └── worker_test.go                              # Тесты
│
│   ├── services                                        # Бизнес-логика (доменный слой)
│   │   ├── metric_get.go                               # Получение одной метрики
│   │   ├── metric_get_mock.go                          # Моки
│   │   ├── metric_get_test.go                          # Тесты
│   │   ├── metric_list_all.go                          # Получение всех метрик
│   │   ├── metric_list_all_mock.go                     # Моки
│   │   ├── metric_list_all_test.go                     # Тесты
│   │   ├── metric_updates.go                           # Пакетное обновление метрик
│   │   ├── metric_updates_mock.go                      # Моки
│   │   └── metric_updates_test.go                      # Тесты
│
│   ├── types                                           # Общие структуры данных проекта
│   │   └── metric.go                                   # Структура и методы для метрик
│
│   └── workers                                         # Фоновые задачи и их обработка
│       ├── metric_agent.go                             # Сбор метрик агентом
│       ├── metric_agent_mock.go                        # Моки
│       ├── metric_agent_test.go                        # Тесты
│       ├── metric_server.go                            # Обработка метрик на стороне сервера
│       ├── metric_server_mock.go                       # Моки
│       └── metric_server_test.go                       # Тесты

├── Makefile                                            # Скрипты сборки, тестирования и запуска
├── migrations
│   └── 20250514024329_create_metrics_table.sql         # SQL-миграция: создание таблицы метрик в БД
└── README.md                                           # Документация по проекту (описание, запуск и т.д.)
```

---

## Используемые технологии

| Технология / Библиотека | Описание                              |
|-------------------------|-------------------------------------|
| Go                      | Язык программирования                |
| Chi                     | HTTP роутер                         |
| Resty                   | HTTP клиент                         |
| Testify                 | Фреймворк для тестирования          |
| Docker                  | Утилита для контейнеризации         |

---

## Инкременты проекта

| Итерация | Описание                                               | 
|----------|--------------------------------------------------------|
| iter1    | Реализация сервера сбора метрик с HTTP API             |
| iter2    | Реализация агента сбора метрик(использован паттерн фасад) |
| iter3    | Добавление к серверу обработчиков дял поулчения метрик |
| iter4    | Добавление флагов для конфигурирования серва и агнта   |
| iter5    | Добавление приоритета конфигураций (env>flag>default)  |
| iter6    | Добавление logging middleware для логирования запросов и ответов сервера  | 
| iter7    | Добавление обновление и получение метрик в теле запроса | 
| iter8    | Добавление gzip middleware для cжатия/разжатия запросов/ответов | 
| iter9    | Добавление загрузки и сохранения метрик сервером из файла | 
| iter10    | Добавление подключения к бд (postgres) | 
| iter11    | Добавление работы сервера с бд и стратегий(использован паттерн стратегия) хранения метрик(бд>файл>память) | 
| iter12    | Добавление обработчика для массового обновления метрик | 
| iter13    | Добавление обработки retriable ошибок сервера и агента | 

---


## Развертывание проекта

1. Клонируйте репозиторий ```git clone git@github.com:sbilibin2017/yp-metrics.git```
2. Скомпилируйте сервер: ```go build -o ./cmd/server/server ./cmd/server/```
3. Скомпилируйте агент: ```go build -o ./cmd/agent/agent ./cmd/agent/```
4. Запустите сервер: ```./cmd/server/server```
5. Запустите агент: ```./cmd/agent/agent```