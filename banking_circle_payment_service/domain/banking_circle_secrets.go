package domain

import (
	"encoding/json"
	"io"
)

type BankingCircleSecret struct {
	Username                          string `json:"username"`
	Password                          string `json:"password"`
	Base64EncodedClientCertPublicKey  string `json:"clientCertPublicKey"`
	Base64EncodedClientCertPrivateKey string `json:"clientCertPrivateKey"`
}

func (bcs BankingCircleSecret) ToJSON() (string, error) {
	bytes, err := json.Marshal(bcs)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func NewBankingCircleSecretFromJSON(secretJSON io.Reader) (BankingCircleSecret, error) {
	var bcSecret BankingCircleSecret
	err := json.NewDecoder(secretJSON).Decode(&bcSecret)
	return bcSecret, err
}
