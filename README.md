# Sequids MVP (gRPC + Protobuf)

## Что сделано в этой версии
- Central ↔ Worker полностью переведены на **gRPC + Protobuf**.
- Удалён legacy пакет binary RPC (`net/rpc`).
- Добавлен операторский CLI: `sequidsctl`.
- Расширен DSL поведения датчиков и аномалий.
- Добавлены RPC для stop/status/logs.

## Быстрый старт

### 1) Запуск central
```bash
go run ./cmd/central serve \
  -grpc-addr :50051 \
  -metrics-addr :8080 \
  -db ./sequids.db \
  -formulas ./configs/formulas/formulas.yaml \
  -anomalies ./configs/anomalies/anomalies.yaml
```

### 2) Запуск worker
```bash
go run ./cmd/worker \
  -id worker-1 \
  -grpc-addr :50052 \
  -metrics-addr :8090 \
  -central-grpc 127.0.0.1:50051 \
  -mqtt-host localhost -mqtt-port 1883 \
  -influx-url http://localhost:8086 \
  -influx-token sequids-token \
  -influx-org sequids \
  -influx-bucket metrics
```

### 3) Управление через CLI
```bash
# старт: загрузка сценария + запуск эксперимента
go run ./cmd/sequidsctl start \
  -grpc 127.0.0.1:50051 \
  -scenario-file ./examples/greenhouse.dsl \
  -scenario-name greenhouse-v2 \
  -seed 42

# статус системы и устройств
go run ./cmd/sequidsctl status -grpc 127.0.0.1:50051

# логи orchestrator
go run ./cmd/sequidsctl logs -grpc 127.0.0.1:50051 -limit 100

# остановка эксперимента
go run ./cmd/sequidsctl stop -grpc 127.0.0.1:50051 -run-id <RUN_ID>
```


### Docker Compose: почему раньше поднималось "не до конца" и что исправлено

Исправлены две причины нестабильного старта в compose:
1. Worker теперь регистрируется в central с retry (раньше при разовом fail завершался).
2. В compose для worker добавлена установка `mosquitto-clients` перед запуском (нужен `mosquitto_pub`).

Поэтому стандартный запуск теперь такой (сначала собери локальные бинарники):
```bash
make build
cd deployments
docker compose up -d
```

Проверка, что всё реально живо:
```bash
docker compose ps
docker compose logs worker --tail=100
```


Если `docker compose up -d` падает с ошибкой `failed to bind host port ... 0.0.0.0:8080 ... address already in use`, значит порт `8080` на хосте уже занят.

Используй другой host-порт для central через переменную окружения:
```bash
cd deployments
CENTRAL_METRICS_PORT=18080 docker compose up -d
```

Если также занят `50051`, можно переопределить и его:
```bash
cd deployments
CENTRAL_METRICS_PORT=18080 CENTRAL_GRPC_PORT=15051 docker compose up -d
```

После этого CLI запускай с новым gRPC-портом, например:
```bash
go run ./cmd/sequidsctl status -grpc 127.0.0.1:15051
```


Если после `docker compose up -d` CLI всё ещё даёт `connection refused` на `127.0.0.1:50051`, проверь:
```bash
cd deployments
docker compose ps
docker compose logs central --tail=100
docker compose logs worker --tail=100
```

В этой версии compose central и worker запускают уже собранные локальные бинарники (`/app/bin/central`, `/app/bin/worker`), поэтому внутри контейнеров больше не нужен `go run` и не требуется доступ к `proxy.golang.org` во время старта. Worker при этом ставит `mosquitto-clients` перед стартом.
Если воркер внутри Docker запущен с дефолтным `-central-grpc localhost:50051`, он автоматически переключается на `central:50051`, чтобы не пытаться подключаться к `::1` внутри собственного контейнера.


Если получаешь ошибку compose вида `services.grafana.environment.[0]: unexpected type map[string]interface {}`, значит в `environment` используется неверный list-формат (например `- KEY: value`).
В этом репозитории используется корректный map-формат:
```yaml
environment:
  GF_SECURITY_ADMIN_USER: admin
  GF_SECURITY_ADMIN_PASSWORD: admin
```

Для multi-worker в Docker обязательно задавай каждому воркеру уникальные `-grpc-addr` и `-advertise-addr` (например `worker:50052`, `worker2:50053`), иначе central может пытаться подключаться к loopback/одному и тому же адресу.

В `deployments/docker-compose.yml` уже добавлен `worker2` с `:50053` и `:8091`, а `telegraf.conf` обновлён для сбора метрик с `worker2:8091`.

Важно: данные device-графиков появятся после запуска эксперимента:
```bash
go run ./cmd/sequidsctl start -grpc 127.0.0.1:50051 -scenario-file ./examples/greenhouse.dsl -scenario-name greenhouse -seed 42
```

## DSL (расширено)

Пример:
```yaml
name: greenhouse-hybrid-v2
devices:
  - id: temp-virt-1
    type: temperature
    topic: iot/greenhouse/temp
    frequency_hz: 1
    formula_ref: temp_daily_sine
    gain: 1.0
    offset: 0.2
    clamp_min: -20
    clamp_max: 60
    jitter_ratio: 0.12
    anomalies:
      - anomaly_ref: mild_sensor_noise
      - anomaly_ref: delayed_delivery

  - id: humidity-virt-1
    type: humidity
    topic: iot/greenhouse/humidity
    frequency_hz: 0.5
    formula_ref: humidity_inverse_wave
    startup_delay_sec: 2
    anomalies:
      - anomaly_ref: intermittent_dropout
      - kind: stuck
        probability: 0.03
        hold_sec: 8
```

### Новые поля поведения устройства
- `gain`, `offset`
- `clamp_min`, `clamp_max`
- `startup_delay_sec`
- `jitter_ratio`

### Новые типы аномалий
- `spike`
- `dropout`
- `stuck` (`hold_sec`)
- `delay` (`duration_sec`)
- также сохранены: `noise`, `false_data`, `drift`


## Grafana дашборды (добавлено)

В проекте добавлены 3 преднастроенных dashboard:
- `Sequids Overview` — общая картина по событиям/ошибкам central+worker.
- `Sequids Device Telemetry` — значения сенсоров, rate сообщений, последние значения по устройствам.
- `Sequids Runs & Throughput` — метрики по `run_id` (пропускная способность и последние сэмплы).

### Где лежат файлы
- JSON: `deployments/grafana/dashboards/*.json`
- provisioning: `deployments/grafana/provisioning/dashboards/dashboards.yaml`

### Как установить/использовать
1. Поднять инфраструктуру:
   ```bash
   cd deployments
   docker compose up -d
   ```
2. Открыть Grafana: `http://localhost:3000` (логин/пароль обычно `admin/admin` на первом входе).
3. Проверить, что datasource `InfluxDB-Sequids` уже создан автоматически.
4. Перейти в папку **Dashboards → Sequids** и открыть нужный dashboard.
5. Запустить эксперимент через CLI (`sequidsctl start`) и наблюдать панели в real-time.


### Что показывает каждый dashboard и каждый блок

#### 1) Sequids Overview
- **Events/sec (service counters)**: скорость роста счётчика `sequids_events_total` (операционная активность сервисов).
  - Как читать: чем выше линия, тем больше событий/сек обрабатывают central/worker.
- **Errors/sec (service counters)**: скорость роста `sequids_errors_total`.
  - Как читать: всплеск вверх = деградация (проблемы publish/influx/rpc и т.п.).
- **Total events (latest)**: последнее абсолютное значение счётчика событий.
  - Как читать: монотонно растёт, полезно для быстрой sanity-проверки «система вообще работает».
- **Total errors (latest)**: последнее абсолютное значение счётчика ошибок.
  - Как читать: если растёт быстрее обычного — нужно открыть логи и сверить с worker publish/influx.
- **Device points/sec**: среднее количество записей `device_metrics` за окно.
  - Как читать: показывает фактическую телеметрию от устройств в Influx.
- **Active devices (last 15m)**: число устройств, которые писали данные за последние 15 минут.
  - Как читать: если ниже ожидаемого количества устройств в сценарии — часть устройств «молчит».

#### 2) Sequids Device Telemetry
- **Device values (virtual + real)**: временные ряды `value` по выбранным `device_id`.
  - Как читать: основная панель качества сигнала; ищем шум, дрейф, выбросы, обрывы.
- **Messages per window by device**: количество сообщений по устройствам в каждом интервале агрегации Grafana.
  - Как читать: просадки до 0 указывают на dropout/остановку run/проблемы публикации.
- **Last values by device/topic/source**: последняя точка по комбинации `device_id/topic/source`.
  - Как читать: быстрый срез «что сейчас последнее пришло» без открытия таймсерий.

#### 3) Sequids Runs & Throughput
- **Points per window by run**: количество точек телеметрии по каждому `run_id` за окно.
  - Как читать: сравнение нагрузки/продуктивности экспериментов между собой.
- **Latest samples by run/device**: последние значения по run и устройствам.
  - Как читать: проверка, что конкретный run ещё жив и какие устройства в нём активны.

### Важные заметки
- Для корректной визуализации `Sequids Runs & Throughput` в Influx теперь пишется tag `run_id` в measurement `device_metrics`.
- Если дашборды не появились, перезапусти Grafana сервис:
  ```bash
  docker compose restart grafana
  ```
- После ручного редактирования JSON в репозитории Grafana подхватит изменения автоматически (интервал обновления provisioning: 30с).

## Что ещё нужно для production-ready прототипа
1. TLS/mTLS и authN/authZ для gRPC.
2. Персистентное хранилище run/log state (сейчас in-memory для runtime/logs).
3. Стриминг логов и телеметрии (server-streaming) вместо только polling.
4. Ретраи/таймауты/circuit breaker для межсервисных вызовов.
5. Явная schema validation DSL (JSONSchema/OpenAPI style).
6. Нагрузочные и chaos-тесты на multi-worker режим.
7. Набор Grafana dashboard/alerts «из коробки».

## Проверка
```bash
go test ./...
go build ./cmd/central ./cmd/worker ./cmd/sequidsctl
```
