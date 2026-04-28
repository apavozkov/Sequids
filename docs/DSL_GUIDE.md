# DSL в Sequids: подробная инструкция

Этот документ описывает DSL сценариев Sequids: какие поля доступны, как их заполнять, как работают формулы и аномалии, и какие есть ограничения реализации.

## 1) Общая структура DSL

Сценарий — это YAML-подобный файл (в проекте используется минимальный line-based парсер).

Минимальный шаблон:

```yaml
name: <имя-сценария>
devices:
  - id: <уникальный-id>
    type: <тип-датчика>
    topic: <mqtt-topic>
    frequency_hz: <частота>
    formula_ref: <имя-формулы-из-каталога>
    anomalies:
      - anomaly_ref: <имя-аномалии-из-каталога>
```

Полный шаблон с расширениями:

```yaml
name: greenhouse-hybrid-v2
devices:
  - id: temp-virt-1
    type: temperature
    topic: iot/greenhouse/temp
    frequency_hz: 1

    # формула: либо formula_ref, либо formula (явное выражение)
    formula_ref: temp_daily_sine
    # formula: "22 + 3*sin(t/120)"

    # пост-обработка сигнала
    gain: 1.0
    offset: 0.2
    clamp_min: -20
    clamp_max: 60
    startup_delay_sec: 1
    jitter_ratio: 0.12

    anomalies:
      - anomaly_ref: mild_sensor_noise
      - anomaly_ref: delayed_delivery
      - kind: stuck
        probability: 0.03
        hold_sec: 8
```

---

## 2) Поля сценария и их смысл

## 2.1 Корневые поля

### `name` (string, обязательно)
- Название сценария.
- Если пустое, парсер вернёт ошибку `scenario name is required`.

### `devices` (list, обязательно)
- Список виртуальных устройств.
- Если список пустой, run формально стартует, но данных не будет.

---

## 2.2 Поля устройства (`devices[]`)

### `id` (string, обязательно)
- Уникальный идентификатор устройства в рамках сценария.
- Используется в payload и метриках.
- Примеры:
  - `temp-virt-1`
  - `humidity-zone-a`

### `type` (string, рекомендуется)
- Логический тип датчика (например, `temperature`, `humidity`, `co2`).
- В runtime не валидируется строго.
- Используется как семантика для человека/каталога формул (`applies_to` — справочное поле).

### `topic` (string, обязательно для полезной работы)
- MQTT topic, куда публикуется телеметрия.
- Примеры:
  - `iot/greenhouse/temp`
  - `factory/line-2/vibration`

### `frequency_hz` (number, > 0 рекомендуется)
- Частота генерации в Гц.
- Интервал = `1 / frequency_hz` секунд.
- Если `<= 0`, runtime автоматически подставит `1`.
- Примеры:
  - `1` → 1 сообщение/сек
  - `0.5` → 1 сообщение/2 сек
  - `10` → 10 сообщений/сек

### `formula_ref` (string) / `formula` (string)
- Источник базовой формулы сигнала (до аномалий).
- Варианты:
  - `formula_ref`: ссылка на выражение из `configs/formulas/formulas.yaml`.
  - `formula`: явное выражение в DSL-устройстве (если задаётся напрямую в структуре).
- На практике в примерах используется `formula_ref`.
- Если указан неверный `formula_ref`, запуск завершится ошибкой resolve (`unknown formula_ref`).

### `gain` (number)
- Мультипликативный коэффициент, применяется к значению формулы.
- Формула преобразования: `value = raw * gain + offset`.
- Особенность: если `gain: 0`, runtime считает это «не задано» и заменяет на `1`.

### `offset` (number)
- Аддитивный сдвиг после `gain`.
- Может быть положительным и отрицательным.

### `clamp_min` (number)
- Нижняя граница итогового значения (после gain/offset).
- Если `value < clamp_min`, берётся `clamp_min`.

### `clamp_max` (number)
- Верхняя граница итогового значения (после gain/offset).
- Если `value > clamp_max`, берётся `clamp_max`.

### `startup_delay_sec` (number >= 0)
- Задержка перед стартом генерации устройства.
- Удобно для staged start или имитации «медленного» подключения.

### `jitter_ratio` (number >= 0)
- Случайный джиттер интервала публикации.
- Реальный интервал = `base_wait * jitter`, где:
  - `base_wait = 1/frequency_hz`
  - `jitter` случайно в диапазоне `[1-jitter_ratio, 1+jitter_ratio]`
- Защита: коэффициент не опускается ниже `0.1`.

### `anomalies` (list)
- Набор аномалий, применяемых к устройству.
- Можно смешивать:
  - ссылочные: `- anomaly_ref: ...`
  - inline: `- kind: ...`

---

## 2.3 Поля аномалий (`anomalies[]`)

Есть 2 способа задать аномалию:

1. Через каталог:
```yaml
- anomaly_ref: mild_sensor_noise
```

2. Inline:
```yaml
- kind: stuck
  probability: 0.03
  hold_sec: 8
```

### Общие поля

### `kind` (string)
Поддерживаемые значения в runtime:
- `noise`
- `false_data`
- `spike`
- `drift`
- `dropout`
- `stuck`
- `delay`

### `probability` (number, обычно 0..1)
- Вероятность срабатывания аномалии на каждом цикле генерации.
- Проверка происходит независимо для каждой аномалии.
- Практически корректно задавать в диапазоне `[0, 1]`.

### Поля по типам аномалий

- `amplitude` (для `noise`, `false_data`, `spike`)
  - Добавляет случайное отклонение: `Δ = U(-amplitude, +amplitude)`.

- `drift_per_sec` (для `drift`)
  - Добавляет тренд: `Δ = t * drift_per_sec`, где `t` — секунды с начала устройства.

- `duration_sec` (для `delay`)
  - Задержка публикации сообщения на указанное время.
  - Если срабатывает несколько delay-анализов за цикл, берётся максимум.

- `hold_sec` (для `stuck`)
  - Заморозка сигнала на последнем значении на заданное число секунд.

---

## 3) Порядок расчёта значения (важно)

На каждом тике для устройства:

1. Вычисляется базовая формула `raw = evalFormula(formula, t)`.
2. Применяется линейная калибровка:
   - `value = raw * gain + offset`.
3. Применяются ограничения `clamp_min/max`.
4. Если ранее активирован `stuck`, значение может быть заменено зафиксированным.
5. Обрабатываются аномалии текущего тика (по вероятностям).
6. Если сработал `dropout` → публикация пропускается.
7. Если сработал `delay` → публикация откладывается.
8. Значение публикуется в MQTT и пишется в Influx.

Примечание: `stuck` фиксирует значение на момент срабатывания, а не исходную формулу.

---

## 4) Формулы: синтаксис и возможности

Формула — арифметическое выражение, где переменная `t` — время в секундах с запуска конкретного устройства.

Поддерживается:
- числа: `10`, `3.14`, `-7`
- операции: `+`, `-`, `*`, `/`
- функции: `sin(...)`, `cos(...)`
- переменная: `t`

Примеры валидных выражений:
- `22 + 3*sin(t/120)`
- `748 + 0.02*t + 0.8*cos(t/200)`
- `1.2 + 0.4*sin(t/8) + 0.1*cos(t/3)`

### Ограничения текущего парсера формул

- Нет скобок общего назначения для группировки, кроме аргументов `sin(...)` и `cos(...)`.
- Нет функций `tan`, `exp`, `log`, `pow`.
- Нет констант (`pi`, `e`).
- Ошибки в выражении в большинстве случаев не валидируются строго, часть некорректного синтаксиса может привести к «молчаливому» 0 в сегменте.

Рекомендация: держать выражения простыми и проверяемыми.

---

## 5) Каталоги формул и аномалий

## 5.1 Формулы (`configs/formulas/formulas.yaml`)

Структура записи:

```yaml
<formula_name>:
  description: "..."
  applies_to: ["temperature"]
  expression: "22 + 3*sin(t/120)"
```

- В runtime реально используются `name` и `expression`.
- `applies_to` сейчас информационное поле (не валидируется при резолве).

## 5.2 Аномалии (`configs/anomalies/anomalies.yaml`)

Структура записи:

```yaml
<anomaly_name>:
  description: "..."
  kind: drift
  probability: 0.15
  drift_per_sec: 0.01
```

Важная особенность текущей реализации каталога:
- При загрузке из `anomaly_ref` гарантированно парсятся поля:
  - `kind`, `probability`, `amplitude`, `drift_per_sec`
- Поля `duration_sec` и `hold_sec` из файла каталога сейчас не загружаются парсером каталога.
- Поэтому для `delay`/`stuck` с параметрами надёжнее задавать inline в самом DSL-устройстве.

---

## 6) Примеры

## 6.1 Температура с мягким шумом и ограничением диапазона

```yaml
name: temp-noise-demo
devices:
  - id: temp-1
    type: temperature
    topic: iot/demo/temp
    frequency_hz: 1
    formula_ref: temp_daily_sine
    gain: 1.0
    offset: 0
    clamp_min: -20
    clamp_max: 60
    anomalies:
      - anomaly_ref: mild_sensor_noise
```

Что получится:
- База: синус вокруг ~22.
- Шум: случайное отклонение.
- Выход не выходит за [-20; 60].

## 6.2 Влажность с задержками и пропусками

```yaml
name: humidity-network-demo
devices:
  - id: hum-1
    type: humidity
    topic: iot/demo/humidity
    frequency_hz: 0.5
    formula_ref: humidity_inverse_wave
    jitter_ratio: 0.2
    anomalies:
      - kind: delay
        probability: 0.2
        duration_sec: 1.3
      - kind: dropout
        probability: 0.05
```

Что получится:
- Сообщения не строго периодические (джиттер).
- Иногда задерживаются.
- Иногда пропускаются полностью.

## 6.3 Inline `stuck` с фиксированием значения

```yaml
name: stuck-demo
devices:
  - id: vib-1
    type: vibration
    topic: iot/demo/vibration
    frequency_hz: 4
    formula: "1.2 + 0.4*sin(t/8)"
    anomalies:
      - kind: stuck
        probability: 0.04
        hold_sec: 5
```

Что получится:
- В норме — колебания вибрации.
- Иногда сигнал «залипает» на 5 секунд.

---

## 7) Практические рекомендации

1. Для production-сценариев держите `probability` в диапазоне `[0,1]`.
2. `frequency_hz` выбирайте исходя из бюджета сообщений и Influx ingestion.
3. Для `delay` и `stuck` лучше inline-описание, если нужны точные `duration_sec`/`hold_sec`.
4. Проверяйте корректность `formula_ref` и `anomaly_ref` заранее.
5. Если нужно «жёсткое» нулевое усиление, используйте формулу `0*... + offset`, потому что `gain: 0` автоматически заменяется на `1` в текущем runtime.

---

## 8) Краткая шпаргалка по полям

- Сценарий:
  - `name`: string, required
  - `devices`: list, required

- Устройство:
  - `id`: string, required
  - `type`: string
  - `topic`: string
  - `frequency_hz`: float (>0 recommended)
  - `formula_ref`: string (catalog key)
  - `formula`: string (inline expression)
  - `gain`: float (0 трактуется как 1)
  - `offset`: float
  - `clamp_min`: float
  - `clamp_max`: float
  - `startup_delay_sec`: float
  - `jitter_ratio`: float
  - `anomalies`: list of anomaly blocks

- Аномалия:
  - `anomaly_ref`: string (catalog key)
  - или inline:
    - `kind`: noise|false_data|spike|drift|dropout|stuck|delay
    - `probability`: float
    - `amplitude`: float
    - `drift_per_sec`: float
    - `duration_sec`: float
    - `hold_sec`: float

---

## 9) Межустройственное взаимодействие через адаптер и in-memory шину

В текущей архитектуре Sequids **виртуальные устройства не общаются с MQTT-брокером напрямую**. Поток данных всегда такой:

1. Воркеры виртуальных устройств публикуют телеметрию **сразу в in-memory data bus**.
2. Адаптеры (MQTT/Zigbee/Bluetooth) читают события из шины и передают их во внешние протоколы/брокеры.
3. Входящие внешние события адаптеры пишут обратно в шину.
4. Из шины события маршрутизируются к получателям (виртуальным и/или реальным устройствам).

Это важно для DSL: в сценарии вы описываете не «прямую отправку в брокер», а правила маршрутизации и реакций внутри шины.

### 9.1 Новые сущности DSL

Для описания связи между устройствами добавляются блоки:

- `flows` — правила передачи данных между источником и приёмником через шину.
- `conditions` — предикаты срабатывания без конструкции `if/else`.
- `actions` — атомарные действия над целевым устройством/топиком.
- `bridges` — имитация протокольных каналов (MQTT/Zigbee/Bluetooth) поверх шины.

Минимальный шаблон:

```yaml
name: device-coupling-demo
devices:
  - id: temp-virt-1
    type: temperature
    formula_ref: temp_daily_sine
  - id: ac-virt-1
    type: hvac

flows:
  - id: temp_to_ac
    from: temp-virt-1
    to: ac-virt-1
    via: adapter.bus
    conditions:
      - metric: value
        op: gt
        threshold: 26
    actions:
      - target: ac-virt-1
        command: power_on
```

### 9.2 Сценарий «датчик температуры → кондиционер» (без `if/else`)

Ниже эквивалент вашей задачи в декларативном стиле:

```yaml
name: temp-ac-threshold
devices:
  - id: temp-virt-1
    type: temperature
    formula_ref: temp_daily_sine
    frequency_hz: 1

  - id: ac-virt-1
    type: hvac

bridges:
  - id: mqtt_bridge_main
    protocol: mqtt
    mode: bus_gateway
    ingress_topic: iot/virt/temp
    egress_topic: iot/virt/ac/cmd

flows:
  - id: sensor_to_broker
    from: temp-virt-1
    to: mqtt_bridge_main
    via: adapter.bus
    actions:
      - command: publish
        payload_field: value

  - id: broker_to_ac
    from: mqtt_bridge_main
    to: ac-virt-1
    via: adapter.bus
    conditions:
      - metric: value
        op: gt
        threshold: 26
    actions:
      - target: ac-virt-1
        command: power_on
```

Что происходит:

- Температурный сенсор генерирует значение.
- Значение попадает в адаптер и шину.
- MQTT bridge читает из шины и публикует во внешний брокер.
- Ответный поток из bridge снова попадает в шину.
- `flow` с условием `op: gt` включает кондиционер.
- Если условие не выполнено, действий нет (это **не** `if/else`, а модель «condition + action»).

### 9.3 Базовые сценарии, которые стоит поддерживать

#### A) Пороговая автоматика (threshold trigger)

Используется для HVAC, сигнализаций, реле:

```yaml
conditions:
  - metric: value
    op: gte
    threshold: 70
actions:
  - command: alarm_on
```

#### B) Оконное условие (range/window)

Действие только при попадании в диапазон:

```yaml
conditions:
  - metric: value
    op: between
    min: 18
    max: 24
actions:
  - command: set_mode_comfort
```

#### C) Антидребезг / подавление флаппинга (hysteresis)

Чтобы устройство не переключалось слишком часто около порога:

```yaml
conditions:
  - metric: value
    op: gt
    threshold: 26
    sustain_sec: 10
actions:
  - command: power_on
```

#### D) Репликация состояния 1→N (fan-out)

Один источник обновляет несколько получателей:

```yaml
flows:
  - id: temp_fanout
    from: temp-virt-1
    via: adapter.bus
    to_many: [ac-virt-1, vent-virt-1, dashboard-proxy]
    actions:
      - command: forward_value
```

#### E) Командная связка «датчик → исполнитель» с cooldown

Защита от повторной команды чаще заданного окна:

```yaml
conditions:
  - metric: value
    op: gt
    threshold: 30
actions:
  - command: power_on
    cooldown_sec: 60
```

### 9.4 Имитация «прямого» общения устройств по Zigbee/Bluetooth

Даже если сценарий логически похож на прямой device-to-device канал, в Sequids он моделируется через шину:

```yaml
bridges:
  - id: zigbee_bridge_lab
    protocol: zigbee
    mode: peer_emulation

  - id: bt_bridge_lab
    protocol: bluetooth
    mode: peer_emulation

flows:
  - id: zigbee_peer_like
    from: motion-virt-1
    to: lamp-virt-1
    via: adapter.bus
    bridge_ref: zigbee_bridge_lab
    conditions:
      - metric: value
        op: eq
        threshold: 1
    actions:
      - command: turn_on

  - id: bt_peer_like
    from: wearable-virt-1
    to: lock-virt-1
    via: adapter.bus
    bridge_ref: bt_bridge_lab
    conditions:
      - metric: rssi
        op: gte
        threshold: -65
    actions:
      - command: unlock
```

Смысл `peer_emulation`:

- поведение похоже на «напрямую по Zigbee/BLE»;
- фактически транспорт и оркестрация остаются внутри adapter + in-memory bus;
- это даёт единый контроль, трассировку и воспроизводимость сценариев.

### 9.5 Рекомендуемые операторы условий

Без `if/else` достаточно стандартного набора операторов:

- `gt`, `gte`, `lt`, `lte`, `eq`, `neq`
- `between`
- `changed` (реакция на изменение)
- `rate_gt` (скорость роста выше порога)
- `sustain_sec` (условие истинно непрерывно N секунд)

Так DSL остаётся декларативным: «когда условие истинно — выполняй action», иначе ничего не делай.

### 9.6 Упрощённая модель обмена через topic + from

Для базовых сценариев используется упрощённая схема:

- У каждого устройства есть `topic` в шине воркера.
- Адаптер воркера автоматически зеркалирует топики виртуальных устройств в MQTT брокер.
- Из MQTT адаптер возвращает в шину данные реальных устройств и других воркеров.

Роли устройств:

- Датчик обычно имеет только `topic` (куда публикует значения).
- Потребитель (например, кондиционер) имеет:
  - `topic` — куда публикует свой статус/состояние;
  - `from` — из какого топика в шине брать входные значения для обработки.

Упрощённый `flows`:

- `id` — идентификатор правила;
- `device` — ID устройства-потребителя из `devices`;
- `conditions` — условия над входным значением, полученным из `device.from`;
- `actions` — действия при истинных условиях.

Поля `from/to/via/bridge_ref` в `flows` больше не требуются для обычных сценариев.

Пример:

```yaml
name: normal-house-temp-ac
devices:
  - id: temp-sensor-1
    type: temperature
    topic: house/livingroom/temperature
    frequency_hz: 1
    formula_ref: temp_daily_sine

  - id: air-conditioner-1
    type: hvac
    topic: house/livingroom/ac/status
    from: house/livingroom/temperature

flows:
  - id: ac_power_on_when_hot
    device: air-conditioner-1
    conditions:
      - metric: value
        op: gt
        threshold: 26
    actions:
      - target: air-conditioner-1
        command: power_on

  - id: ac_power_off_when_cool
    device: air-conditioner-1
    conditions:
      - metric: value
        op: lt
        threshold: 24
    actions:
      - target: air-conditioner-1
        command: power_off
```
