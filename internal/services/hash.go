package services

import "golang.org/x/crypto/bcrypt"

type HashProvider interface {
	GetHash(password string) ([]byte, error)
	CompareHash(password string, hash []byte) bool
}

func NewBcryptHashProvider() *BcryptHashProvider {
	return &BcryptHashProvider{}
}

type BcryptHashProvider struct {
}

func (p *BcryptHashProvider) GetHash(password string) ([]byte, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return hashedPassword, nil
}

func (p *BcryptHashProvider) CompareHash(password string, hash []byte) bool {
	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	return err == nil
}
