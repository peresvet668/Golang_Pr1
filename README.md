# Wallet API

REST API на Go и PostgreSQL. Docker Compose запускает приложение и базу; тестовый кошелёк создаётся автоматически.

## Запуск

Нужны Docker Compose и GNU Make. Для Go-тестов на хосте нужен Go 1.27.3.

```sh
git clone https://github.com/peresvet668/Golang_Pr1.git
cd Golang_Pr1
cp config.env.example config.env
make up
```

`make up` работает в текущем терминале. Команды ниже выполняй во втором терминале из папки проекта.

## Основные команды

| Команда | Действие |
| --- | --- |
| `make up` | Собрать и запустить API с PostgreSQL |
| `make balance` | Получить баланс кошелька |
| `make deposit AMOUNT=100` | Пополнить баланс на 100 |
| `make withdraw AMOUNT=50` | Снять 50 со счёта |
| `make down` | Остановить контейнеры, сохранив данные БД |

По умолчанию используется кошелёк `11111111-1111-4111-8111-111111111111` и API `http://localhost:8080`. Их можно переопределить:

```sh
make balance WALLET_ID=YOUR_UUID API_URL=http://localhost:8080
```

## Вспомогательные команды

| Команда | Действие |
| --- | --- |
| `make help` | Показать список команд; также запускается командой `make` |
| `make build` | Собрать Docker-образ API |
| `make logs` | Следить за логами API и PostgreSQL |
| `make test` | Запустить быстрые Go-тесты |
| `make test-race` | Запустить тесты с race detector |
| `make db-up` | Запустить только PostgreSQL |
| `make test-integration` | Проверить 1000 параллельных пополнений и снятий через HTTP и PostgreSQL |
| `make test-all` | Запустить быстрые и интеграционные тесты |
| `make benchmark` | Измерить скорость конкурентных HTTP-операций |

Тесты и benchmark требуют Go и используют PostgreSQL на `localhost:5432`.

## HTTP API

- `POST /api/v1/wallet` — пополнение (`DEPOSIT`) или снятие (`WITHDRAW`); поля JSON: `walletId`, `operationType`, `amount`.
- `GET /api/v1/wallets/{WALLET_UUID}` — получить баланс.
- `GET /healthz` — проверить готовность API.

Изменение баланса выполняется атомарно в PostgreSQL; отрицательный баланс не допускается.
