package postgresql

import "github.com/saltpay/settlements-payments-system/internal/domain/models"

type failureEventFromPG struct {
	Details errorDetails `json:"details"`
}

func (c failureEventFromPG) GetCode() models.DomainFailureReasonCode {
	return models.DomainFailureReasonCode(c.Details.FailureReason.Code + c.Details.RejectedReason.Code)
}

type code struct {
	Code string `json:"code"`
}

type errorDetails struct {
	RejectedReason code `json:"rejectionReason"`
	FailureReason  code `json:"failureReason"`
}
