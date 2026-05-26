# Trending Search Service

Сервис реального времени, который показывает топ поисковых запросов за последние 5 минут.
Читает события из Kafka, хранит всё in-memory, отдаёт топ за микросекунды.

---

## Быстрый старт

### 1. Клонируй репозиторий

```bash
git clone https://github.com/zamurabims/kafkatest.git
cd kafkatest
go mod tidy
```

### 2. Запусти

```bash
docker-compose up --build
```

Дождись пока оба контейнера поднимутся. Сервис будет доступен на `http://localhost:8080`.

Проверь что всё живое:
```bash
curl http://localhost:8080/health
```

---

## Как отправить события в Kafka

> ⚠️ **Важно**: timestamp должен быть актуальным — сервис хранит только события за последние 5 минут. Примеры ниже подставляют текущее время автоматически.

### macOS / Linux

```bash
echo "{\"query\":\"iphone 16\",\"session_id\":\"s1\",\"timestamp\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}" | \
  docker exec -i $(docker ps -qf "name=kafka") \
  kafka-console-producer --bootstrap-server localhost:9092 --topic search-events
```

> ⚠️ **macOS**: синтаксис `<<<` не работает в zsh — используй `echo ... | docker exec -i` как показано выше.

**Отправить несколько событий сразу:**
```bash
for q in "iphone 16" "iphone 16" "nike sneakers" "nike sneakers" "nike sneakers" "adidas" "macbook pro"; do
  echo "{\"query\":\"$q\",\"session_id\":\"s$RANDOM\",\"timestamp\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}" | \
  docker exec -i $(docker ps -qf "name=kafka") \
  kafka-console-producer --bootstrap-server localhost:9092 --topic search-events
done
```

### Windows (PowerShell)

```powershell
$ts = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$msg = "{`"query`":`"iphone 16`",`"session_id`":`"s1`",`"timestamp`":`"$ts`"}"
echo $msg | docker exec -i $(docker ps -qf "name=kafka") `
  kafka-console-producer --bootstrap-server localhost:9092 --topic search-events
```

**Отправить несколько событий:**
```powershell
$queries = @("iphone 16", "iphone 16", "nike sneakers", "nike sneakers", "nike sneakers", "adidas", "macbook pro")
foreach ($q in $queries) {
  $ts = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
  $msg = "{`"query`":`"$q`",`"session_id`":`"s$($RANDOM % 1000)`",`"timestamp`":`"$ts`"}"
  echo $msg | docker exec -i $(docker ps -qf "name=kafka") `
    kafka-console-producer --bootstrap-server localhost:9092 --topic search-events
}
```

### Узнать точное имя kafka-контейнера

Если команды выше не работают — узнай имя контейнера явно и подставь его:

```bash
docker ps --format "table {{.Names}}"
```

Пример вывода: `kafkatest-kafka-1`. Тогда используй:
```bash
docker exec -i kafkatest-kafka-1 kafka-console-producer ...
```

---

## API

### Получить топ запросов

```bash
curl "http://localhost:8080/top?limit=5"
```

Ответ:
```json
[
  {"query":"nike sneakers","count":3},
  {"query":"iphone 16","count":2},
  {"query":"adidas","count":1},
  {"query":"macbook pro","count":1}
]
```

Параметр `limit` — сколько запросов вернуть (по умолчанию 10).

---

### Стоп-лист

Слова из стоп-листа мгновенно исчезают из топа без перезапуска сервиса.

**Добавить слово:**
```bash
curl -X POST http://localhost:8080/stoplist \
  -H "Content-Type: application/json" \
  -d '{"word":"nike sneakers"}'
```

**Удалить слово:**
```bash
curl -X DELETE "http://localhost:8080/stoplist/nike%20sneakers"
```

**Посмотреть стоп-лист:**
```bash
curl http://localhost:8080/stoplist
```

---

### Прочее

```bash
curl http://localhost:8080/health    # healthcheck
curl http://localhost:8080/metrics   # Prometheus метрики
```

---

## Проверка защиты от накруток

Бот шлёт один и тот же запрос 5 раз с одной сессией — должен засчитаться как 1.

**macOS / Linux:**
```bash
for i in 1 2 3 4 5; do
  echo "{\"query\":\"spam\",\"session_id\":\"bot-1\",\"timestamp\":\"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"}" | \
  docker exec -i $(docker ps -qf "name=kafka") \
  kafka-console-producer --bootstrap-server localhost:9092 --topic search-events
done

curl "http://localhost:8080/top?limit=10"
# spam будет count=1, а не count=5
```

**Windows (PowerShell):**
```powershell
1..5 | ForEach-Object {
  $ts = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
  $msg = "{`"query`":`"spam`",`"session_id`":`"bot-1`",`"timestamp`":`"$ts`"}"
  echo $msg | docker exec -i $(docker ps -qf "name=kafka") `
    kafka-console-producer --bootstrap-server localhost:9092 --topic search-events
}

curl "http://localhost:8080/top?limit=10"
```

---

## Если что-то не работает

**Проверить логи сервиса:**
```bash
docker logs $(docker ps -qf "name=trending") 2>&1 | tail -20
```

**Проверить что сообщения дошли до Kafka:**
```bash
docker exec $(docker ps -qf "name=kafka") \
  kafka-console-consumer --bootstrap-server localhost:9092 \
  --topic search-events --from-beginning --max-messages 5
```

**Топ пустой хотя сообщения есть** — скорее всего timestamp в событии устарел. Убедись что используешь актуальное время, а не хардкод вроде `2025-01-01T00:00:00Z`.

**Полный перезапуск с очисткой данных:**
```bash
docker-compose down -v
docker-compose up --build
```

---

## Контракт данных (Kafka payload)

```json
{
  "query": "nike sneakers",
  "session_id": "user-session-abc123",
  "timestamp": "2026-05-26T10:00:00Z"
}
```

| Поле | Тип | Зачем |
|---|---|---|
| `query` | string | Поисковый запрос. Нормализуется на стороне сервиса (lowercase, trim). |
| `session_id` | string | Идентификатор сессии. Один источник засчитывается максимум один раз за бакет (10 сек) для одного запроса — защита от накруток. |
| `timestamp` | RFC3339 | Время события на стороне отправителя. Используем его, а не время получения — защита от сетевых задержек и лагов Kafka. |

---

## Архитектура

### Как устроено хранение

**Скользящее окно из бакетов**

5 минут разбиты на 30 бакетов по 10 секунд. Каждый бакет хранит:
```
map[query] → set<session_id>
```

Set вместо счётчика даёт дедупликацию бесплатно — бот с одной сессией даст `count=1` в бакете независимо от числа запросов.

**Кешированный топ**

Фоновый воркер каждые 500мс:
1. Суммирует все живые бакеты
2. Сортирует по count
3. Атомарно заменяет кеш (`atomic.StorePointer`)

Чтение топа — `atomic.LoadPointer` без локов. При highload (10-50x чтений vs записей) это критично.

**Стоп-лист**

`map[string]struct{}` + `RWMutex`. Фильтрация при чтении из кеша — слово исчезает из топа сразу после добавления в стоп-лист.

### Поток данных

```
Kafka → Consumer → normalize → stoplist check → Window.Record()
                                                       ↓
                                               bucket[slot][query].add(session_id)

Background worker (500ms) → RebuildCache() → atomic.Store(cachedTop)

GET /top → atomic.Load(cachedTop) → filter(stoplist) → top-N → JSON
```

---

## Trade-offs

| Проблема | Решение |
|---|---|
| Что считать "популярным"? | Уникальные session_id — честнее отражает реальный интерес |
| Как бороться с накрутками? | Дедупликация по session_id внутри бакета |
| Точность vs скорость | Кеш обновляется раз в 500мс — для виджета на главной приемлемо |
| Персистентность | Намеренно отсутствует — при рестарте через 5 минут данные актуальны |
| Большое число уникальных сессий | Можно заменить set на HyperLogLog — но это усложнение без реальной проблемы (YAGNI) |

---

## Нагрузочное тестирование

```bash
# Установить hey
go install github.com/rakyll/hey@latest

# Залить данные
python3 scripts/gen_events.py --count 5000 --rate 1000

# Запустить тест
chmod +x scripts/load_test.sh
./scripts/load_test.sh
```

Read path не использует мьютексы — `atomic.LoadPointer` + итерация по срезу.
При 200 конкурентных клиентах latency p99 < 2ms.

---

## Тесты

```bash
go test ./...
```

Покрыты: запись, дедупликация, eviction, фильтрация стоп-листа, лимит топа.

---

## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `HTTP_ADDR` | `:8080` | Адрес HTTP сервера |
| `KAFKA_BROKERS` | `localhost:9092` | Kafka brokers (через запятую) |
| `KAFKA_TOPIC` | `search-events` | Топик Kafka |
| `KAFKA_GROUP_ID` | `trending-service` | Consumer group ID |
| `BUCKET_COUNT` | `30` | Количество временных бакетов |
| `BUCKET_DURATION_SEC` | `10` | Длительность бакета в секундах |
| `CACHE_REFRESH_MS` | `500` | Интервал пересчёта кеша в мс |

---

## Prometheus метрики

| Метрика | Тип | Описание |
|---|---|---|
| `trending_records_total` | Counter | Всего обработанных событий |
| `trending_records_dropped_total{reason}` | Counter | Отброшенные события (invalid / stoplist) |
| `trending_top_request_duration_seconds` | Histogram | Latency GET /top |
| `trending_cache_rebuild_duration_seconds` | Histogram | Время пересчёта кеша |
| `trending_cache_size` | Gauge | Уникальных запросов в кеше |
| `trending_stoplist_size` | Gauge | Слов в стоп-листе |

```bash
curl http://localhost:8080/metrics | grep trending_
```
