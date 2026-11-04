# Wallet API

Небольшое REST-приложение на Go и PostgreSQL. Docker Compose запускает два сервиса: API и базу данных.

## Запуск

```sh
cp config.env.example config.env
docker compose up --build
```

API будет доступно на `http://localhost:8080`. При первом запуске приложение создаёт таблицу `wallets` и один кошелёк с нулевым балансом:

```text
11111111-1111-4111-8111-111111111111
```

Настоящий `config.env` не хранится в Git. Для локального запуска используйте `config.env.example` как основу.

## API

### Пополнить или списать средства

`POST /api/v1/wallet`

```json
{
  "walletId": "11111111-1111-4111-8111-111111111111",
  "operationType": "DEPOSIT",
  "amount": 1000
}
```

`operationType` принимает `DEPOSIT` или `WITHDRAW`. Сумма должна быть положительным целым числом. В задании поле ID написано как `valletId`; API принимает и это написание, и обычное `walletId`.

Успешный ответ:

```json
{
  "walletId": "11111111-1111-4111-8111-111111111111",
  "balance": 1000
}
```

Недостаточный баланс возвращает `409 Conflict`, неизвестный кошелёк — `404 Not Found`, некорректный запрос — `400 Bad Request`.

### Получить баланс

`GET /api/v1/wallets/{WALLET_UUID}`

```sh
curl http://localhost:8080/api/v1/wallets/11111111-1111-4111-8111-111111111111
```

## Как устроена конкурентная запись

HTTP-обработчик проверяет запрос и передаёт его сервису. PostgreSQL-репозиторий меняет баланс одним условным `UPDATE`: снятие выполняется только при достаточном балансе, а пополнение — пока оно не переполнит `BIGINT`. Поэтому параллельные запросы к одной строке не теряют изменения и не уводят баланс в минус. Защита находится в базе, а не в локальном Go-мьютексе.

## Тесты

Быстрые unit-тесты:

```sh
go test ./...
```

PostgreSQL integration-тест отправляет по 1000 параллельных пополнений и списаний через HTTP и проверяет итоговый баланс. Он пропускается, если `TEST_DATABASE_URL` не задан. Compose публикует PostgreSQL только на loopback-порт `5432`, поэтому локально можно выполнить:

```powershell
$env:TEST_DATABASE_URL = "postgres://wallet:local_wallet_password@localhost:5432/wallet?sslmode=disable"
go test ./...
```

Этот тест проверяет корректность при параллельной работе. Фактические 1000 RPS нужно оценивать отдельным нагрузочным запуском на целевом оборудовании.

Для измерения пропускной способности на той же машине запустите benchmark:

```powershell
go test -run '^$' -bench BenchmarkConcurrentWalletWrites -benchtime=10s ./internal/repository/postgres
```

В выводе будет метрика `requests/s`; benchmark также требует заданный `TEST_DATABASE_URL`.
