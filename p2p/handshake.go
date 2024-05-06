package p2p

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log/slog"
)

type HandshakeFunc func(*TCPPeer, GameVariant, string) error

type HandshakePass struct {
	GameVariant GameVariant
	Version     string
}

func PerformHandshake(p *TCPPeer, variant GameVariant, version string) error {
	slog.Info("Performing Handshake")
	hp := &HandshakePass{}
	if err := gob.NewDecoder(p.conn).Decode(hp); err != nil {
		return err
	}

	slog.Info("Received Handshake", "GameVariant", hp.GameVariant, "Version", hp.Version)
	if hp.GameVariant != variant {
		slog.Error("Game Variant Mismatch", "Expected", variant, "Received", hp.GameVariant)
		return errors.ErrUnsupported
	}

	if hp.Version != version {
		slog.Error("Version Mismatch", "Expected", version, "Received", hp.Version)
		return errors.ErrUnsupported
	}

	slog.Info("Handshake Performed", "GameVariant", hp.GameVariant, "Version", hp.Version)
	return nil
}

func SendHandshake(p *TCPPeer, variant GameVariant, version string) error {
	hp := &HandshakePass{
		GameVariant: variant,
		Version:     version,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hp); err != nil {
		return err
	}

	return p.Send(buf.Bytes())
}
