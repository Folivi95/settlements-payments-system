package http_client

import (
	"encoding/json"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
)

type BankingCirclePaymentStatusResponse struct {
	Status ports.PaymentStatus `json:"status"`
}

func (b BankingCirclePaymentStatusResponse) ToJSON() ([]byte, error) {
	return json.Marshal(b)
}

func NewBankingCirclePaymentStatusResponseFromJSON(in []byte) (BankingCirclePaymentStatusResponse, error) {
	var out BankingCirclePaymentStatusResponse
	err := json.Unmarshal(in, &out)
	return out, err
}
