package models

import (
	"encoding/json"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type AccountBalance struct {
	Result []Balance `json:"result"`
}

type Balance struct {
	Currency         models.CurrencyCode `json:"currency"`
	BeginOfDayAmount float64             `json:"beginOfDayAmount"`
	IntraDayAmount   float64             `json:"intraDayAmount"`
}

func (a AccountBalance) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

func NewAccountBalanceFromJSON(in []byte) (AccountBalance, error) {
	var out AccountBalance
	err := json.Unmarshal(in, &out)
	return out, err
}
