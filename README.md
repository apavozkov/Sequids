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
