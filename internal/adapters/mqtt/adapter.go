package mqtt

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

type MosquittoAdapter struct {
	Host string
	Port int
}

func (m MosquittoAdapter) Publish(ctx context.Context, topic string, payload []byte) error {
	cmd := exec.CommandContext(ctx, "mosquitto_pub", "-h", m.Host, "-p", fmt.Sprint(m.Port), "-t", topic, "-m", string(payload))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mosquitto_pub: %w: %s", err, stderr.String())
	}
	return nil
}
