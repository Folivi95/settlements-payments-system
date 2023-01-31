package models

import (
	"encoding/json"
	"io"
)

type PaymentCurrencyReport struct {
	CurrencyReport map[string]CurrencyStats `json:"currency_report"`
}

type CurrencyStats struct {
	Successful       uint    `json:"successful"`
	Failures         uint    `json:"failures"`
	Total            uint    `json:"total"`
	SuccessfulAmount float64 `json:"successful_amount"`
	FailuresAmount   float64 `json:"failures_amount"`
	TotalAmount      float64 `json:"total_amount"`
}

func NewPaymentCurrencyReportFromJSON(in io.Reader) (PaymentCurrencyReport, error) {
	var out PaymentCurrencyReport
	err := json.NewDecoder(in).Decode(&out)
	return out, err
}
