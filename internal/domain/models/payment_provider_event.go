package models

import (
	"encoding/json"
	"errors"
	"time"
)

// PaymentProviderEvent is a generic payment event that is raised by Banking Circle (can be by other providers too).
type PaymentProviderEvent struct {
	CreatedOn                time.Time                `json:"createdOn"`
	Type                     PaymentProviderEventType `json:"type"`
	PaymentInstruction       PaymentInstruction       `json:"paymentInstruction"`
	PaymentProviderName      PaymentProviderName      `json:"paymentProviderName"`
	PaymentProviderPaymentID ProviderPaymentID        `json:"paymentProviderPaymentId"`
	BankingReference         BankingReference         `json:"bankingReference"`
	FailureReason            FailureReason            `json:"failureReason"`
}

func (e PaymentProviderEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func NewPaymentProviderEventFromJSON(in []byte) (PaymentProviderEvent, error) {
	var out PaymentProviderEvent
	err := json.Unmarshal(in, &out)
	return out, err
}

type FailureReason struct {
	Code    PPEventFailureCode `json:"code"`
	Message string             `json:"message"`
}

func (f FailureReason) IsEmpty() bool {
	return f.Code == "" && f.Message == ""
}

type PPEventFailureCode string

const (
	RejectedCode                   PPEventFailureCode = "REJECTED"
	StuckInPending                 PPEventFailureCode = "STUCK_IN_PENDING"
	UnhandledPaymentProviderStatus PPEventFailureCode = "UNHANDLED_PAYMENT_PROVIDER_STATUS"
	TransportFailure               PPEventFailureCode = "TRANSPORT_FAILURE"
	NoSourceAccount                PPEventFailureCode = "NO_SOURCE_ACCOUNT"
	MissingFunding                 PPEventFailureCode = "MISSING_FUNDING"
)

type PaymentProviderEventType string

const (
	Submitted PaymentProviderEventType = "SUBMITTED"
	Processed PaymentProviderEventType = "PROCESSED"
	Failure   PaymentProviderEventType = "FAILURE"
)

type (
	PaymentProviderName string
	ProviderPaymentID   string
	BankingReference    string
)

const (
	BC PaymentProviderName = "banking_circle"
)

func NewPaymentProviderEvent(
	now time.Time,
	eventType PaymentProviderEventType,
	paymentInstruction PaymentInstruction,
	paymentProviderName PaymentProviderName,
	paymentProviderPaymentID ProviderPaymentID,
	bankingReference BankingReference,
	failure *FailureReason,
) (PaymentProviderEvent, error) {
	if paymentInstruction.ID() == "" {
		return PaymentProviderEvent{}, errors.New("paymentInstructionId is required and can not be empty string")
	}

	if paymentProviderName == "" {
		return PaymentProviderEvent{}, errors.New("paymentProviderName is required and can not be empty string")
	}

	if eventType == Failure {
		if failure == nil {
			return PaymentProviderEvent{}, errors.New("FailureReason object can not be nil")
		}

		if (*failure == FailureReason{}) {
			return PaymentProviderEvent{}, errors.New("FailureReason object can not be empty")
		}
	}

	event := PaymentProviderEvent{
		CreatedOn:                now,
		Type:                     eventType,
		PaymentInstruction:       paymentInstruction,
		PaymentProviderName:      paymentProviderName,
		PaymentProviderPaymentID: paymentProviderPaymentID,
		BankingReference:         bankingReference,
	}

	if eventType == Failure {
		event.FailureReason = *failure
	}

	return event, nil
}
