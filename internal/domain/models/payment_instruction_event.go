package models

import (
	"time"
)

type PaymentInstructionEventType string

type PaymentProviderType string

const (
	BankingCircle PaymentProviderType = "banking_circle"
	Islandsbanki  PaymentProviderType = "islandsbanki"
)

const (
	DomainReceived                   PaymentInstructionEventType = "DOMAIN.RECEIVED"
	DomainSubmittedToPaymentProvider PaymentInstructionEventType = "DOMAIN.SUBMITTED_TO_PAYMENT_PROVIDER"
	DomainProcessingSucceeded        PaymentInstructionEventType = "DOMAIN.PROCESSING_SUCCEEDED"
	DomainProcessingFailed           PaymentInstructionEventType = "DOMAIN.PROCESSING_FAILED"
	DomainRejected                   PaymentInstructionEventType = "DOMAIN.REJECTED"
)

type PaymentInstructionEvent struct {
	Type      PaymentInstructionEventType `json:"type"`
	CreatedOn time.Time                   `json:"createdOn"`
	Details   interface{}                 `json:"details"`
}

type DomainSubmittedToPaymentProviderEventDetails struct {
	PaymentProviderType PaymentProviderType `json:"paymentProviderType"`
}

type DomainProcessingSucceededEventDetails struct {
	PaymentProviderPaymentID ProviderPaymentID `json:"paymentProviderPaymentID"`
	// inconsistent naming convention? everything else is paymentproviderID
}

type DomainProcessingFailedEventDetails struct {
	FailureReason     PIFailureReason   `json:"failureReason"`
	PaymentProviderID ProviderPaymentID `json:"paymentProviderPaymentID"`
}

type DomainRejectedFailedEventDetails struct {
	FailureReason     PIFailureReason   `json:"failureReason"`
	PaymentProviderID ProviderPaymentID `json:"paymentProviderPaymentID"`
}

type DomainUnhandledStatusFailedEventDetails struct {
	FailureReason     PIFailureReason   `json:"failureReason"`
	PaymentProviderID ProviderPaymentID `json:"paymentProviderPaymentID"`
}

type DomainTransportErrorEventDetails struct {
	FailureReason     PIFailureReason   `json:"failureReason"`
	PaymentProviderID ProviderPaymentID `json:"paymentProviderPaymentID"`
}

type DomainRejectedEventDetails struct {
	FailureReason PIFailureReason `json:"rejectionReason"`
}

type DomainFailureReasonCode string

const (
	StuckPayment     DomainFailureReasonCode = "STUCK_IN_PENDING"
	RejectedPayment  DomainFailureReasonCode = "REJECTED_PAYMENT"
	Unhandled        DomainFailureReasonCode = "UNHANDLED"
	TransportMishap  DomainFailureReasonCode = "TRANSPORT_MISHAP"
	NoSourceAcct     DomainFailureReasonCode = "NO_SOURCE_ACCOUNT"
	FailedValidation DomainFailureReasonCode = "FAILED_VALIDATION"
	MissingFunds     DomainFailureReasonCode = "MISSING_FUNDS"
)

type PIFailureReason struct {
	Code    DomainFailureReasonCode `json:"code"`
	Message string                  `json:"message"`
}
