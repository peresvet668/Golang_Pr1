# Wallet API

Небольшой REST-сервис на Go с PostgreSQL. Docker Compose запускает API и базу данных; один тестовый кошелёк создаётся при старте приложения.

## Быстрый старт

Нужны Docker Engine с Compose plugin и GNU Make. Для Go-тестов на хосте также нужен Go 1.27.3.

```sh
git clone https://github.com/peresvet668/Golang_Pr1.git
cd Golang_Pr1
cp config.env.example config.env
make up
```

API слушает `http://localhost:8080`. UUID готового кошелька:

```text
11111111-1111-4111-8111-111111111111
```

`config.env` содержит локальные настройки и игнорируется Git. Не используйте примерные пароль и строку подключения для production.

## HTTP API

### Изменить баланс

`POST /api/v1/wallet`

```json
{
  "walletId": "11111111-1111-4111-8111-111111111111",
  "operationType": "DEPOSIT",
  "amount": 1000
}
```

`operationType`: `DEPOSIT` или `WITHDRAW`. `amount` — положительное целое число. В PDF поле написано с опечаткой `valletId`; API принимает и `valletId`, и стандартное `walletId`.

Успешный ответ содержит новый баланс:

```json
{
  "walletId": "11111111-1111-4111-8111-111111111111",
  "balance": 1000
}
```

Недостаточно средств — `409`, неизвестный кошелёк — `404`, некорректный запрос — `400`.

### Получить баланс

`GET /api/v1/wallets/{WALLET_UUID}`

```sh
curl http://localhost:8080/api/v1/wallets/11111111-1111-4111-8111-111111111111
```

## Makefile

Запусти `make help`, чтобы увидеть краткую справку. Полный список:

| Команда | Что делает |
| --- | --- |
| `make help` | Показывает основные команды. Также используется по умолчанию при запуске `make`. |
| `make up` | Собирает образ API и запускает API вместе с PostgreSQL в foreground. |
| `make down` | Останавливает Compose-сервисы. Данные в именованном volume сохраняются. |
| `make build` | Собирает только Docker-образ API. |
| `make logs` | Показывает логи API и PostgreSQL в реальном времени. |
| `make test` | Запускает Go-тесты без race detector. Интеграционный тест пропускается, если `TEST_DATABASE_URL` не задан. |
| `make test-race` | Запускает тесты с Go race detector. Интеграционный тест запускается, если задан `TEST_DATABASE_URL`. |
| `make db-up` | Запускает только PostgreSQL и ждёт состояния healthy. |
| `make test-integration` | Запускает PostgreSQL, затем с race detector прогоняет HTTP-тест: 1000 параллельных пополнений и 1000 снятий с проверкой итогового баланса. |
| `make test-all` | Последовательно запускает быстрые и интеграционные тесты. |
| `make benchmark` | Запускает 10-секундный конкурентный benchmark через HTTP и PostgreSQL; выводит `requests/s`. |

Интеграционные тесты и benchmark используют PostgreSQL на `localhost:5432`. Переопределить строку подключения можно так:

```sh
make test-integration LOCAL_TEST_DATABASE_URL="postgres://wallet:local_wallet_password@localhost:5432/wallet?sslmode=disable"
```

Benchmark показывает результат на текущем оборудовании; он помогает оценить целевые 1000 RPS, но не гарантирует такую скорость на другой машине.

## Конкурентные обновления

Пополнение и снятие выполняются атомарным условным `UPDATE` в PostgreSQL. Это не теряет одновременные изменения и не позволяет балансу стать отрицательным. Интеграционный тест проверяет корректность параллельных запросов; benchmark измеряет пропускную способность.