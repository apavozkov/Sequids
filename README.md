# Sequids MVP (engineering VKR stand)

## Ключевые изменения в этой версии
1. **DSL теперь хранит только имена формул и аномалий** (`formula_ref`, `anomaly_ref`).
2. **Каталоги формул/аномалий вынесены в отдельные файлы**:
   - `configs/formulas/formulas.yaml`
   - `configs/anomalies/anomalies.yaml`
3. **Из orchestration flow удалён HTTP JSON API** (больше нет `/workers/register`, `/experiments/run` и т.п. по HTTP).
4. Межкомпонентное взаимодействие central↔worker работает через **binary RPC транспорт** (`net/rpc`, TCP) как промежуточный этап.
5. Метрики central/worker снимаются Telegraf-ом и пишутся в InfluxDB; device-метрики пишутся Listener-модулем воркера в InfluxDB.

> В репозитории сохранён protobuf-контракт (`api/proto/orchestrator.proto`). Ниже есть пошаговый migration-план, как довести текущую версию до **чистого gRPC+Protobuf**.

---

## Архитектура
### Central
- RPC сервис `OrchestratorService`:
  - `RegisterWorker`
  - `Heartbeat`
  - `PushScenario`
  - `RunExperiment`
- Выбор мастер-воркера по текущей нагрузке.
- Назначение роли воркерам (`AssignRole`) и запуск нагрузок (`StartWorkload`).
- Хранение сценариев в SQLite.

### Worker
- RPC сервис `WorkerService`:
  - `AssignRole`
  - `StartWorkload`
- Эмуляция устройств с формулами и аномалиями.
- In-memory шина данных + Listener на запись в InfluxDB.

### Логика Listener
- Если воркер мастер: пишет все локальные datapoint (виртуальные и реальные).
- Если воркер не мастер: пишет только локальные datapoint от виртуальных датчиков.

---

## DSL (подробно)

### Пример сценария
```yaml
name: greenhouse-hybrid
devices:
  - id: temp-virt-1
    type: temperature
    topic: iot/greenhouse/temp
    frequency_hz: 1
    formula_ref: temp_daily_sine
    anomalies:
      - anomaly_ref: mild_sensor_noise
      - anomaly_ref: slow_positive_drift
```

### Поля DSL
#### Уровень сценария
- `name` — имя сценария.
- `devices` — список устройств.

#### Устройство
- `id` — идентификатор устройства.
- `type` — тип датчика (temperature, humidity...).
- `topic` — MQTT topic для публикации.
- `frequency_hz` — частота генерации.
- `formula_ref` — ссылка на формулу из каталога.
- `anomalies` -> `anomaly_ref` — ссылки на профили аномалий из каталога.

### Ограничения DSL (MVP)
- Нет `if/when`.
- Нет встроенных циклов/ветвлений.
- Нет сложной schema-validation до запуска.

---

## Каталог формул (`configs/formulas/formulas.yaml`)
В файле уже добавлены примеры:
- `temp_daily_sine` — температура с плавной синусоидой.
- `humidity_inverse_wave` — влажность с обратной корреляцией к температуре.
- `pressure_micro_drift` — давление с трендом + косинус.
- `co2_occupancy_cycle` — циклическое изменение CO2 в присутствии людей.
- `vibration_machine_periodic` — периодическая вибрация оборудования.

Каждая формула содержит:
- `description` — что моделирует;
- `applies_to` — типы датчиков;
- `expression` — математическое выражение.

---

## Каталог аномалий (`configs/anomalies/anomalies.yaml`)
Добавлены готовые профили:
- `mild_sensor_noise` — слабый шум.
- `severe_false_data_spike` — редкие ложные выбросы.
- `slow_positive_drift` — медленный положительный дрейф.
- `intermittent_negative_drift` — периодический отрицательный дрейф.

Каждый профиль содержит:
- `description`;
- `kind` (`noise|false_data|drift`);
- `probability`;
- `amplitude` и/или `drift_per_sec`.

---

## Развёртывание инфраструктуры (Grafana/Telegraf/InfluxDB/MQTT)

Перед запуском локального демо на хосте установите MQTT CLI клиент:
```bash
sudo apt update && sudo apt install -y mosquitto-clients
```
(нужен бинарник `mosquitto_pub`).

## 1) MQTT broker (Mosquitto)
Файл: `deployments/mosquitto.conf`.
- Для MVP включён anonymous доступ и listener 1883.
- В production включить auth + ACL + TLS.

## 2) InfluxDB
В `docker-compose` используется авто-инициализация:
- org: `sequids`
- bucket: `metrics`
- token: `sequids-token`

## 3) Telegraf
Файл: `deployments/telegraf.conf`.
- `inputs.prometheus` читает:
  - `http://central:8080/metrics`
  - `http://worker:8090/metrics`
- `outputs.influxdb_v2` пишет в Influx bucket `metrics`.

## 4) Grafana
Datasource provisioned автоматически (`deployments/grafana/provisioning/datasources/datasource.yaml`).
После старта:
1. открыть `http://localhost:3000`
2. войти (admin/admin при первом входе обычно через UI Grafana)
3. проверить datasource `InfluxDB-Sequids`
4. строить панели по measurement:
   - технические (`sequids_events_total`, `sequids_errors_total`)
   - device (`device_metrics`)

## 5) Запуск стенда
```bash
cd deployments
docker compose up -d
```

---

## Запуск central/worker вручную
### Central
```bash
go run ./cmd/central serve \
  -rpc-addr :50051 \
  -metrics-addr :8080 \
  -db ./sequids.db \
  -formulas ./configs/formulas/formulas.yaml \
  -anomalies ./configs/anomalies/anomalies.yaml
```

### Worker
```bash
go run ./cmd/worker \
  -rpc-addr :50052 \
  -metrics-addr :8090 \
  -central-rpc 127.0.0.1:50051 \
  -mqtt-host localhost -mqtt-port 1883 \
  -influx-url http://localhost:8086 \
  -influx-token sequids-token \
  -influx-org sequids \
  -influx-bucket metrics
```

### Загрузка сценария и запуск эксперимента
```bash
go run ./cmd/central push-scenario -rpc 127.0.0.1:50051 -file ./examples/greenhouse.dsl -name greenhouse
go run ./cmd/central run -rpc 127.0.0.1:50051 -scenario <SCENARIO_ID> -seed 42
```

---


## Частые ошибки запуска и как исправить
- `mosquitto_pub not found` или `publish failed ... executable file not found`:
  1. Установите клиент: `sudo apt update && sudo apt install -y mosquitto-clients`
  2. Проверьте: `which mosquitto_pub`
- `address already in use` для `:50051` или `:50052`:
  1. Остановите старые процессы: `./scripts/stop_demo.sh`
  2. Либо запустите с другими портами: `-rpc-addr :50151` / `-rpc-addr :50152`
- `scenario <id> not found`:
  1. Убедитесь, что сценарий был загружен именно в этот central (`-rpc <host:port>`).
  2. Убедитесь, что central запущен с тем же `-db`, куда вы ранее пушили сценарий.

## Как довести до чистого gRPC+Protobuf (без legacy RPC)
Сейчас транспорт бинарный RPC (TCP), чтобы убрать HTTP JSON и сохранить работоспособность в этой среде.

Чтобы перейти на **полный gRPC+Protobuf**:
1. Установить инструменты:
   - `protoc`
   - `protoc-gen-go`
   - `protoc-gen-go-grpc`
2. Сгенерировать Go-код из `api/proto/orchestrator.proto`.
3. Добавить зависимости в `go.mod`:
   - `google.golang.org/grpc`
   - `google.golang.org/protobuf`
4. Заменить `net/rpc` сервер/клиенты в:
   - `cmd/central/main.go`
   - `cmd/worker/main.go`
   - `internal/orchestrator/server.go`
   на gRPC server/client (unary RPC).
5. Оставить текущую доменную логику без изменений:
   - балансировка мастер-воркера;
   - резолв `formula_ref`/`anomaly_ref`;
   - runtime + listener.
6. После migration удалить пакет `internal/transport/rpc`.

Итог: транспорт станет строго gRPC+Protobuf, без промежуточных RPC-слоёв.

---

## Проверка
```bash
go test ./...
go build ./cmd/central ./cmd/worker
```
