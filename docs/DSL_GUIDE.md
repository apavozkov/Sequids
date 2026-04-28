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
