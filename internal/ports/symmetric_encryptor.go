package ports

type SymmetricEncryptor interface {
	Encrypt(plaintext []byte, key []byte) ([]byte, error)
	Decrypt(ciphertext []byte, key []byte) ([]byte, error)
	CreateKey() ([]byte, error)
}
