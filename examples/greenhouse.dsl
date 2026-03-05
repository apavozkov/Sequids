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
  - id: humidity-virt-1
    type: humidity
    topic: iot/greenhouse/humidity
    frequency_hz: 0.5
    formula_ref: humidity_inverse_wave
    anomalies:
      - anomaly_ref: severe_false_data_spike
