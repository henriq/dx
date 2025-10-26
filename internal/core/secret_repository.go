package core

import (
	"dx/internal/core/domain"
	"dx/internal/ports"
	"encoding/json"
	"fmt"
	"os"
)

type SecretsRepository interface {
	LoadSecrets(configContextName string) ([]*domain.Secret, error)
	SaveSecrets(secrets []*domain.Secret, configContextName string) error
}

func ProvideEncryptedFileSecretRepository(
	fileSystem ports.FileSystem,
	keyring ports.Keyring,
	encryptor ports.SymmetricEncryptor,
) SecretsRepository {
	return &EncryptedFileSecretRepository{
		fileSystem: fileSystem,
		keyring:    keyring,
		encryptor:  encryptor,
	}
}

type EncryptedFileSecretRepository struct {
	fileSystem ports.FileSystem
	keyring    ports.Keyring
	encryptor  ports.SymmetricEncryptor
}

func (e EncryptedFileSecretRepository) LoadSecrets(configContextName string) ([]*domain.Secret, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	secretsFilePath := fmt.Sprintf("%s/.dx/%s/secrets", homeDir, configContextName)
	secretFileExists, err := e.fileSystem.FileExists(secretsFilePath)
	if err != nil {
		return nil, err
	}
	keyExists, err := e.keyring.HasKey(fmt.Sprintf("%s-encryption-key", configContextName))
	if err != nil {
		return nil, err
	}
	if !secretFileExists || !keyExists {
		return []*domain.Secret{}, nil
	}

	encryptedSecrets, err := e.fileSystem.ReadFile(secretsFilePath)
	if err != nil {
		return nil, err
	}

	key, err := e.keyring.GetKey(fmt.Sprintf("%s-encryption-key", configContextName))
	if err != nil {
		return nil, err
	}

	decryptedSecrets, err := e.encryptor.Decrypt(encryptedSecrets, []byte(key))

	if err != nil {
		return nil, err
	}

	var secrets []*domain.Secret
	err = json.Unmarshal(decryptedSecrets, &secrets)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

func (e EncryptedFileSecretRepository) SaveSecrets(
	secrets []*domain.Secret,
	configContextName string,
) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	secretsFilePath := fmt.Sprintf("%s/.dx/%s/secrets", homeDir, configContextName)
	keyExists, err := e.keyring.HasKey(fmt.Sprintf("%s-encryption-key", configContextName))
	if err != nil {
		return err
	}
	if !keyExists {
		key, err := e.encryptor.CreateKey()
		if err != nil {
			return err
		}
		err = e.keyring.SetKey(fmt.Sprintf("%s-encryption-key", configContextName), string(key))
		if err != nil {
			return err
		}
	}
	key, err := e.keyring.GetKey(fmt.Sprintf("%s-encryption-key", configContextName))
	if err != nil {
		return err
	}

	secretBytes, err := json.Marshal(secrets)
	if err != nil {
		return err
	}

	encryptedSecrets, err := e.encryptor.Encrypt(secretBytes, []byte(key))
	if err != nil {
		return err
	}

	err = e.fileSystem.WriteFile(secretsFilePath, encryptedSecrets, ports.ReadWrite)
	if err != nil {
		return err
	}

	return nil
}
