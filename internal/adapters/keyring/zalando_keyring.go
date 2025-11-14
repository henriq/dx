package keyring

import (
	"dx/internal/ports"
	"errors"

	"github.com/zalando/go-keyring"
)

type ZalandoKeyring struct{}

func ProvideZalandoKeyring() ports.Keyring {
	return ZalandoKeyring{}
}

func (z ZalandoKeyring) GetKey(keyName string) (string, error) {
	return keyring.Get("se.henriq.dx", keyName)
}

func (z ZalandoKeyring) SetKey(keyName string, keyValue string) error {
	return keyring.Set("se.henriq.dx", keyName, keyValue)
}

func (z ZalandoKeyring) HasKey(keyName string) (bool, error) {
	_, err := keyring.Get("se.henriq.dx", keyName)
	if errors.Is(err, keyring.ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}
