# Sequids MVP Prototype

Ниже — пошаговая инструкция, как развернуть и запустить минимальный прототип на Ubuntu 24.04.

## 1. Установите зависимости

### 1.1 Go

```bash
sudo apt update
sudo apt install -y golang-go
```

Проверьте установку:

```bash
go version
```

### 1.2 Protobuf compiler (protoc)

```bash
sudo apt install -y protobuf-compiler
```

Проверьте:

```bash
protoc --version
```

### 1.3 gRPC-плагины для Go

Установите плагины генерации кода:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
```

Убедитесь, что `$GOPATH/bin` в вашем `PATH` (обычно `$HOME/go/bin`):

```bash
export PATH="$PATH:$HOME/go/bin"
```

Чтобы сделать это постоянным, добавьте строку выше в `~/.bashrc`.

## 2. Склонируйте проект и подготовьте зависимости

```bash
git clone <URL_ВАШЕГО_РЕПО>
cd Sequids
```

Скачайте зависимости Go:

```bash
go mod tidy
```

## 3. (Опционально) Перегенерируйте protobuf-код

В репозитории уже лежат Go-стабы для gRPC, но при необходимости вы можете
перегенерировать их вручную:

```bash
protoc \
  --go_out=. \
  --go-grpc_out=. \
  proto/orchestrator.proto
```

## 4. Запуск воркера

Воркера можно запускать на отдельной машине. Для локального запуска:

```bash
go run ./cmd/worker \
  -listen :9100 \
  -worker-id worker-1 \
  -orchestrator-addr localhost:9000
```

## 5. Запуск оркестратора

В отдельном терминале:

```bash
go run ./cmd/orchestrator \
  -listen :9000 \
  -worker-addr localhost:9100 \
  -sensor-id sensor-1 \
  -worker-id worker-1 \
  -interval-ms 1000
```

После запуска оркестратор отправит воркеру команду создать датчик. В терминале
воркера будут появляться строки вида:

```
sensor=sensor-1 value=42.1234
```

## 6. Запуск на разных машинах

1. На машине воркера запустите:
   ```bash
   go run ./cmd/worker -listen :9100 -worker-id worker-remote -orchestrator-addr <IP_ОРКЕСТРАТОРА>:9000
   ```
2. На машине оркестратора:
   ```bash
   go run ./cmd/orchestrator -listen :9000 -worker-addr <IP_ВОРКЕРА>:9100 -sensor-id sensor-remote -worker-id worker-remote -interval-ms 1000
   ```

Убедитесь, что порты 9000 и 9100 доступны через firewall (ufw/iptables).

## 7. Остановка

Обе программы корректно завершаются по `Ctrl+C`.

## 8. Частая ошибка с `no matching versions for query "latest"`

Если видите ошибку:

```text
go: finding module for package github.com/<you>/sequids/pkg/proto/orchestratorpb
go: ... no matching versions for query "latest"
```

это значит, что импорт указывает на несуществующий пакет
`pkg/proto/orchestratorpb`.

Правильно в этом проекте импортировать пакет так:

```go
orchestratorpb "github.com/<you>/sequids/pkg/proto"
```

Проверьте 3 вещи:

1. В `go.mod`:
   ```go
   module github.com/<you>/sequids
   ```
2. Во всех `import` в `cmd/*` и `internal/*`:
   `github.com/<you>/sequids/pkg/proto` (без `/orchestratorpb` в конце).
3. В `proto/orchestrator.proto`:
   ```proto
   option go_package = "github.com/<you>/sequids/pkg/proto;orchestratorpb";
   ```

После исправления:

```bash
go mod tidy
```

Если репозиторий приватный, дополнительно:

```bash
go env -w GOPRIVATE=github.com/<you>/*
```

## 9. Ошибка `panic: non-contiguous repeated field`

Причина: ранее в проекте использовался вручную собранный protobuf-descriptor в
`pkg/proto/orchestrator.pb.go`, который мог быть несовместим с вашей версией
`google.golang.org/protobuf`.

Исправление в текущей версии:
- protobuf runtime descriptor убран из рантайма MVP;
- gRPC обмен выполняется через JSON codec (поверх gRPC), поэтому паника больше
  не возникает;
- API и структуры сообщений (`CreateSensorRequest`, `ReportSensorStatusRequest` и
  т.д.) сохранены.

После обновления репозитория выполните:

```bash
go mod tidy
go run ./cmd/worker -listen :9100 -worker-id worker-1 -orchestrator-addr localhost:9000
```

В другом терминале:

```bash
go run ./cmd/orchestrator -listen :9000 -worker-addr localhost:9100 -sensor-id sensor-1 -worker-id worker-1 -interval-ms 1000
```
