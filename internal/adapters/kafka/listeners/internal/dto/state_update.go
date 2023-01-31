package dto

import (
	"time"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type State string

const (
	StateProcessed State = "PROCESSED"
	StateFailure   State = "FAILURE"
	StateSubmitted State = "SUBMITTED"
)

type FailureCode string

const (
	FailureCodeRejectedFailureCode FailureCode = "REJECTED"
	FailureStuckInPending          FailureCode = "STUCK_IN_PENDING"
	FailureCodeMissingFunding      FailureCode = "MISSING_FUNDING"
	FailureTransportFailure        FailureCode = "TRANSPORT_FAILURE"
	FailureUnexpectedReason        FailureCode = "UNHANDLED_FAILURE_REASON"
)

type FailureReason struct {
	Code    FailureCode `json:"code"`
	Message string      `json:"message"`
}

type PaymentStateUpdate struct {
	PaymentInstructionID string        `json:"payment_instruction_id"`
	UpdatedState         State         `json:"updated_state"`
	FailureReason        FailureReason `json:"failure_reason,omitempty"`
}

func (psu PaymentStateUpdate) PaymentInstructionStatus() models.PaymentInstructionStatus {
	switch psu.UpdatedState {
	case StateProcessed:
		return models.Successful
	case StateFailure:
		return models.Failed
	case StateSubmitted:
		return models.StateSubmitted
	default:
		return ""
	}
}

func (psu PaymentStateUpdate) Event() models.PaymentInstructionEvent {
	event := models.PaymentInstructionEvent{
		CreatedOn: time.Now(),
		Details:   psu.FailureReason,
	}

	switch psu.UpdatedState {
	case StateProcessed:
		event.Type = models.DomainProcessingSucceeded
	case StateFailure:
		event.Type = models.DomainRejected
	}

	return event
}
