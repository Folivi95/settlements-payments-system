package models

import (
	"encoding/json"
	"io"
)

type PaymentReport struct {
	Stats          PaymentStats   `json:"stats"`
	FailedPayments FailedPayments `json:"failed_payments"`
}

type PaymentStats struct {
	Successful             uint `json:"successful"`
	Failed                 uint `json:"failed"`
	SubmittedForProcessing uint `json:"submitted_for_processing"`
	Rejected               uint `json:"rejected"`
	Received               uint `json:"received"`
}

type FailedPayments struct {
	FailedStats        FailedStats         `json:"failed_stats"`
	FailedInstructions []FailedInstruction `json:"failed_instructions"`
}

type FailedStats struct {
	Rejected         uint `json:"rejected"`
	StuckInPending   uint `json:"stuck_in_pending"`
	Unhandled        uint `json:"unhandled"`
	TransportMishap  uint `json:"transport_mishap"`
	NoSourceAcct     uint `json:"no_source_acct"`
	FailedValidation uint `json:"failed_validation"`
	MissingFunds     uint `json:"missing_funds"`
}

type FailedInstruction struct {
	ID       PaymentInstructionID    `json:"id"`
	Currency CurrencyCode            `json:"currency"`
	Mid      string                  `json:"mid"`
	Reason   DomainFailureReasonCode `json:"reason"`
}

func NewPaymentReportFromJSON(in io.Reader) (PaymentReport, error) {
	var out PaymentReport
	err := json.NewDecoder(in).Decode(&out)
	return out, err
}
