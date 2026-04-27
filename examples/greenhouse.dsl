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
      - anomaly_ref: slow_positive_drift
      - anomaly_ref: delayed_delivery

  - id: humidity-virt-1
    type: humidity
    topic: iot/greenhouse/humidity
    frequency_hz: 0.5
    formula_ref: humidity_inverse_wave
    startup_delay_sec: 2
    anomalies:
      - anomaly_ref: severe_false_data_spike
      - anomaly_ref: intermittent_dropout
      - kind: stuck
        probability: 0.03
        hold_sec: 8
