## Быстрый запуск

Нужен файл `.env` в корне проекта. Минимальный пример:

```env
APP_HTTP_PORT=8080
APP_JWT_SECRET=replace-with-long-random-secret
APP_JWT_TTL_MINUTES=60

POSTGRES_DB=room_booking
POSTGRES_USER=admin
POSTGRES_PASSWORD=replace-with-strong-password
POSTGRES_PORT=5432

APP_POSTGRES_DSN=postgres://admin:replace-with-strong-password@localhost:5432/room_booking?sslmode=disable
APP_POSTGRES_TEST_DSN=postgres://postgres:postgres@localhost:55432/room_booking_test?sslmode=disable
```

Где `APP_POSTGRES_DSN` должен ссылаться на ту же основную БД, что описана через `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` и `POSTGRES_PORT`. То есть данные внутри URL в `APP_POSTGRES_DSN` должны соответствовать этим же env-переменным.

`APP_POSTGRES_TEST_DSN` менее критичен: он нужен только для integration tests и может указывать на отдельную test DB с любыми подходящими тестовыми параметрами.

Запуск через сервиса через Makefile:

```bash
make up
```
Тестовые данные можно создать через makefile:

```bash
make seed
```

После старта сервис доступен на `http://localhost:8080`.

Какие контейнеры поднимутся:
- `postgres`
- `migrate` — накатывает `migrations/000001_init.up.sql`
- `app`

Чтобы положить сервер с сохранением данных в бд на следующие запуски:
```bash
make down
```
Чтобы положить сервер с удалением данных в бд на следующие запуски:
```bash
make down-v
```

## Тесты
```bash
make test
```

На данный момент не кросспакетный coverage проекта  44.9%, а кросспакетный 68.6%.

`make test` использует отдельный одноразовый контейнер `postgres:17` для интеграционных тестов. Он сам поднимает test DB на `localhost:55432`, прогоняет тесты и затем удаляет контейнер.

## Основные решения

### Архитектура

Проект разделен на слои:
- `internal/domain` — сущности и инварианты
- `internal/application` — use cases и порты
- `internal/infrastructure` — HTTP, Postgres, JWT, bcrypt, clock, id generator
- `cmd/server` — composition root

Так бизнес-правила не смешиваются с транспортом и БД, а use case'ы тестируются отдельно.

### Генерация слотов

Выбрана комбинированная стратегия `eager + lazy`:
- при создании расписания слоты генерируются на ближайшие 7 дней вперед
- если пользователь запрашивает дату, на которую слоты еще не были созданы, они лениво генерируются при первом запросе
- горизонт запросов ограничен 30 днями вперед

Почему так:
- по условию 99.9% запросов приходятся на ближайшие 7 дней
- это ускоряет самый горячий эндпоинт `/rooms/{roomId}/slots/list`
- при этом сервис не ломается на датах чуть дальше горячего окна

### Ограничения консистентности

Ключевые ограничения дублируются на уровне БД:
- одно расписание на комнату
- уникальный слот по `room_id + start_at`
- только одна активная бронь на слот через partial unique index

Это защищает от гонок не только в коде, но и в PostgreSQL.

## Решение по `conferenceLink`

Дополнительное задание с `createConferenceLink` реализовано через порт `ConferenceLinkService` и мок-адаптер `internal/infrastructure/conference/mock.go`.

Текущее поведение:
- если `createConferenceLink=false`, бронь создается как обычно
- если `createConferenceLink=true`, сервис сначала запрашивает ссылку у внешнего `Conference Service`, затем создает бронь и сохраняет ссылку в записи брони
- в проекте используется мок, который возвращает детерминированную ссылку вида `https://conference.local/rooms/<bookingId>`

### Что решили делать при сбоях

#### 1. Внешний сервис недоступен или возвращает ошибку до ответа

Решение: бронь не создается, клиент получает ошибку.

Почему так:
- пользователь явно запросил бронь со ссылкой на конференцию
- silent fallback к брони без ссылки делал бы поведение неочевидным
- так сохраняется простая и честная семантика: либо бронь создана вместе со ссылкой, либо операция целиком не выполнена

#### 2. Ошибка после успешного ответа внешнего сервиса, но до сохранения брони в БД

Текущее решение: запрос завершается ошибкой, запись брони в БД не появляется, а внешний side effect считается допустимым остаточным эффектом.

Почему так:
- в учебном проекте внешний сервис замокан и не хранит собственного состояния
- полноценная компенсация потребовала бы отдельного контракта удаления конференции или outbox/retry-механизма
- для тестового задания это избыточно, поэтому выбран простой и явный компромисс

Что делали бы в production-версии:
- либо добавили бы компенсирующую операцию удаления конференции
- либо вынесли бы интеграцию во внешний сервис в outbox/asynchronous flow с ретраями и reconciliation

## Что покрыто тестами

Unit tests покрывают:
- доменные инварианты
- application use cases
- config loader
- password hashing
- JWT issue/parse

Integration tests покрывают:
- сценарий `создание переговорки -> создание расписания -> создание брони`
- сценарий `отмена брони`
- auth flow `register -> login`
- основные HTTP-маршруты и role-based access

## Полезные команды

Остановить сервисы:

```bash
docker compose down
```

Перезапустить с пересборкой:

```bash
docker compose down
docker compose up --build
```


## Нагрузочное тестирование

Для нагрузки используется отдельный perf-контур:
- отдельный контейнер Postgres `room-booking-perf-db` на `localhost:55433`
- отдельный экземпляр приложения на `localhost:8081`
- те же миграции, накатанные через `psql`
- тот же сид, но с override `APP_POSTGRES_DSN`

Основная команда:

```bash
make perf-test-slots
```

Перед первым запуском нужен установленный `k6` на хосте. Например, на Ubuntu/WSL можно поставить так:

```bash
sudo snap install k6
```

Без `k6` сама нагрузка не запустится, хотя perf DB и perf app targets останутся доступны отдельно.

Что она делает:
- поднимает perf Postgres
- накатывает `migrations/000001_init.up.sql`
- наполняет perf БД сидом
- запускает отдельный app instance
- гоняет host-side `k6` по hot endpoint `/rooms/{roomId}/slots/list`
- после завершения останавливает perf app и удаляет perf DB container

По умолчанию сценарий гоняется с параметрами:
- `VUS=50`
- `DURATION=30s`

Их можно переопределить так:

```bash
LOAD_VUS=100 LOAD_DURATION=60s make perf-test-slots
```

Краткий результат локального прогона hot endpoint `/rooms/{roomId}/slots/list`:
- инструмент: `k6`
- профиль нагрузки: `50 VUs`, `30s`
- `http_req_failed = 0.00%`
- `avg = 12.47ms`
- `p95 = 16.12ms`
- ориентир из задания `p95 < 200ms` выполнен с большим запасом
