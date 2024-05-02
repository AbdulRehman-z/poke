package p2p

import (
	"io"
	"log/slog"
)

type Handler interface {
	HandleMessage(msg *Message) error
}

type DefaultHandler struct{}

func (h *DefaultHandler) HandleMessage(msg *Message) error {
	b, err := io.ReadAll(msg.Payload)
	if err != nil {
		return err
	}

	slog.Info("Received message", "from", msg.From, "to", msg.To, "payload", string(b))
	return nil
}
