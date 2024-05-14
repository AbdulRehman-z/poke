package deck

import (
	"bytes"
	"encoding/gob"
)

func Encrypt(key, payload []byte) ([]byte, error) {
	encOutput := make([]byte, len(payload))
	for i := 0; i < len(payload); i++ {
		encOutput[i] = payload[i] ^ key[i%len(key)]
	}

	return encOutput, nil
}

func EncryptCard(key []byte, card Card) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(card); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecryptCard(key, encCard []byte) (*Card, error) {
	card := &Card{}

	b, err := Encrypt(key, encCard)
	if err != nil {
		return nil, err
	}

	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(card); err != nil {
		return nil, err
	}

	return card, nil
}
