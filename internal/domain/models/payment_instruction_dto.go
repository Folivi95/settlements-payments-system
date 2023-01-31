package models

type PaymentInstructionDTO struct {
	IncomingInstruction IncomingInstruction       `json:"incomingInstruction"`
	ID                  PaymentInstructionID      `json:"id"`
	Version             int                       `json:"version"`
	PaymentProvider     PaymentProviderType       `json:"paymentProvider"`
	Status              PaymentInstructionStatus  `json:"status"`
	Events              []PaymentInstructionEvent `json:"events"`
}

func NewPaymentInstructionDTO(paymentInstruction PaymentInstruction) PaymentInstructionDTO {
	return PaymentInstructionDTO{
		IncomingInstruction: paymentInstruction.IncomingInstruction,
		ID:                  paymentInstruction.id,
		Version:             paymentInstruction.version,
		PaymentProvider:     paymentInstruction.paymentProvider,
		Status:              paymentInstruction.status,
		Events:              paymentInstruction.events,
	}
}
