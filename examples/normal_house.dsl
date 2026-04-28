name: normal-house-temp-ac

# HVAC status convention used by runtime/Grafana:
# state_code: 0 = OFF, 1 = ON
# state_text: "off" / "on"

devices:
  - id: temp-sensor-1
    type: temperature
    topic: house/livingroom/temperature
    frequency_hz: 1
    formula_ref: temp_daily_sine
    gain: 1.0
    offset: 0

  - id: air-conditioner-1
    type: hvac
    # Device publishes current AC status here.
    topic: house/livingroom/ac/status
    # Consumer input topic (temperature stream).
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
