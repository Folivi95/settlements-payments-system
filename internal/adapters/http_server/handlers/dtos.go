package handlers

import (
	"encoding/json"
	"io"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type PaymentStatusResponse struct {
	Status models.PaymentInstructionStatus `json:"status"`
}

type PaymentResponse struct {
	ID models.ProviderPaymentID `json:"id"`
}

func NewPaymentResponseFromJSON(in io.Reader) (PaymentResponse, error) {
	var out PaymentResponse
	err := json.NewDecoder(in).Decode(&out)
	return out, err
}

type DLQURLs []string
