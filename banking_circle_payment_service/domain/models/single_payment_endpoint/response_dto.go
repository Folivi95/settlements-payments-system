package single_payment_endpoint

import (
	"encoding/json"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type ResponseDto struct {
	PaymentID        models.ProviderPaymentID `json:"paymentId"`
	Status           string                   `json:"status"`
	BankingReference models.BankingReference  `json:"bankingReference"`
}

func (r ResponseDto) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func NewResponseDTOFromJSON(in []byte, bankingReference models.BankingReference) (ResponseDto, error) {
	var out ResponseDto
	err := json.Unmarshal(in, &out)
	out.BankingReference = bankingReference
	return out, err
}
