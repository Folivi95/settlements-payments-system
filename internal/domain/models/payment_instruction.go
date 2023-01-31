package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PaymentInstruction is a request to make a settlements payment to a merchant.
type PaymentInstruction struct {
	IncomingInstruction IncomingInstruction
	id                  PaymentInstructionID
	version             int
	paymentProvider     PaymentProviderType
	status              PaymentInstructionStatus
	events              []PaymentInstructionEvent
}

type (
	PaymentInstructionStatus string
	PaymentInstructionID     string
)

const (
	Received               PaymentInstructionStatus = "RECEIVED"
	Rejected               PaymentInstructionStatus = "REJECTED"
	SubmittedForProcessing PaymentInstructionStatus = "SUBMITTED_FOR_PROCESSING"
	Successful             PaymentInstructionStatus = "PROCESSING_SUCCEEDED"
	Failed                 PaymentInstructionStatus = "PROCESSING_FAILED"
	StateSubmitted         PaymentInstructionStatus = "SUBMITTED"
)

func NewPaymentInstruction(instruction IncomingInstruction) PaymentInstruction {
	instruction = instruction.NormaliseAccountNumber()

	return PaymentInstruction{
		status:              Received,
		id:                  PaymentInstructionID(uuid.NewString()),
		version:             1,
		paymentProvider:     "",
		IncomingInstruction: instruction,
		events: []PaymentInstructionEvent{
			{
				Type:      DomainReceived,
				CreatedOn: time.Now(),
			},
		},
	}
}

func (p *PaymentInstruction) UnmarshalJSON(in []byte) error {
	pi, err := NewPaymentInstructionFromJSON(in)
	if err != nil {
		return err
	}

	// todo 2 is there a way to avoid mapping these manually? or at least reduce duplication
	p.IncomingInstruction = pi.IncomingInstruction
	p.id = pi.id
	p.status = pi.status
	p.paymentProvider = pi.paymentProvider
	p.events = pi.events
	p.version = pi.version

	return nil
}

func NewPaymentInstructionFromJSON(in []byte) (PaymentInstruction, error) {
	var instruction PaymentInstructionDTO
	err := json.Unmarshal(in, &instruction)
	if err != nil {
		return PaymentInstruction{}, err
	}
	return PaymentInstruction{
		IncomingInstruction: instruction.IncomingInstruction,
		id:                  instruction.ID,
		version:             instruction.Version,
		paymentProvider:     instruction.PaymentProvider,
		status:              instruction.Status,
		events:              instruction.Events,
	}, nil
}

var _ json.Marshaler = PaymentInstruction{}

func (p PaymentInstruction) MarshalJSON() ([]byte, error) {
	return json.Marshal(NewPaymentInstructionDTO(p))
}

func (p PaymentInstruction) MustToJSON() ([]byte, error) {
	serialised, err := p.MarshalJSON()
	if err != nil {
		return []byte{}, nil
	}
	return serialised, nil
}

func (p *PaymentInstruction) Rejected(instruction IncomingInstruction, err string) {
	p.updateStatus(Rejected)
	instructionJSON, _ := instruction.ToJSON()
	p.events = append(p.events, PaymentInstructionEvent{
		Type:      DomainRejected,
		CreatedOn: time.Now(),
		Details: DomainRejectedEventDetails{FailureReason: PIFailureReason{
			Code: FailedValidation,
			Message: InvalidPaymentInstructionError{
				UnderlyingError: errors.New(err),
				InstructionJSON: string(instructionJSON),
			}.Error(),
		}},
	})
}

func (p *PaymentInstruction) SubmitForProcessing() {
	p.updateStatus(SubmittedForProcessing)
	p.events = append(p.events, PaymentInstructionEvent{
		Type:      DomainSubmittedToPaymentProvider,
		CreatedOn: time.Now(),
		Details:   DomainSubmittedToPaymentProviderEventDetails{PaymentProviderType: p.paymentProvider},
	})
}

func (p *PaymentInstruction) RouteToPaymentProvider() {
	switch {
	case strings.HasPrefix(p.IncomingInstruction.Metadata.Sender, "ISB"):
		p.SetPaymentProvider(Islandsbanki)
	case strings.HasPrefix(p.IncomingInstruction.Metadata.Sender, "RB"):
		p.SetPaymentProvider(Islandsbanki)
	case strings.HasPrefix(p.IncomingInstruction.Metadata.Sender, "SAXO"):
		p.SetPaymentProvider(BankingCircle)
	case p.IncomingInstruction.Metadata.Sender == "":
		if p.IncomingInstruction.Metadata.Source == "Solanteq" && p.IncomingInstruction.Payment.Currency.IsoCode == "ISK" {
			p.SetPaymentProvider(Islandsbanki)
		} else {
			p.SetPaymentProvider(BankingCircle)
		}
	}
}

func (p PaymentInstruction) GetStatus() PaymentInstructionStatus {
	return p.status
}

func (p *PaymentInstruction) SetStatus(status PaymentInstructionStatus) {
	p.updateStatus(status)
}

func (p PaymentInstruction) Version() int {
	return p.version
}

func (p *PaymentInstruction) PaymentProvider() PaymentProviderType {
	return p.paymentProvider
}

func (p PaymentInstruction) Events() []PaymentInstructionEvent {
	return p.events
}

func (p PaymentInstruction) ID() PaymentInstructionID {
	return p.id
}

func (p PaymentInstruction) ContractNumber() string {
	return p.IncomingInstruction.Merchant.ContractNumber
}

func (p *PaymentInstruction) SetSourceAccount(account string) {
	p.IncomingInstruction.Payment.Sender.AccountNumber = account
}

func (p *PaymentInstruction) SetPaymentProvider(providerType PaymentProviderType) {
	p.paymentProvider = providerType
}

func (p PaymentInstruction) BusinessID() string {
	return fmt.Sprintf(
		"%s#%s#%s",
		p.IncomingInstruction.Merchant.ContractNumber,
		p.IncomingInstruction.Payment.ExecutionDate.Format("2006-01-02"),
		p.IncomingInstruction.Payment.Amount,
	)
}

func (p *PaymentInstruction) TrackPPEvent(event PaymentProviderEvent) {
	if event.Type == Failure {
		p.failedToProcessPayment(event)
	} else {
		p.successfullyProcessed(event.PaymentProviderPaymentID)
	}
}

func (p *PaymentInstruction) updateStatus(newStatus PaymentInstructionStatus) {
	p.status = newStatus
	p.version += 1
}

func (p *PaymentInstruction) failedToProcessPayment(event PaymentProviderEvent) {
	p.updateStatus(Failed)

	var instructionEvent PaymentInstructionEvent
	id := event.PaymentProviderPaymentID
	switch event.FailureReason.Code {
	case RejectedCode:
		instructionEvent = PaymentInstructionEvent{
			Type:      DomainProcessingFailed,
			CreatedOn: time.Now(),
			Details: DomainRejectedFailedEventDetails{
				FailureReason: PIFailureReason{
					Code:    RejectedPayment,
					Message: event.FailureReason.Message,
				},
				PaymentProviderID: id,
			},
		}
	case StuckInPending:
		instructionEvent = PaymentInstructionEvent{
			Type:      DomainProcessingFailed,
			CreatedOn: time.Now(),
			Details: DomainProcessingFailedEventDetails{
				FailureReason: PIFailureReason{
					Code:    StuckPayment,
					Message: event.FailureReason.Message,
				},
				PaymentProviderID: id,
			},
		}
	case TransportFailure:
		instructionEvent = PaymentInstructionEvent{
			Type:      DomainProcessingFailed,
			CreatedOn: time.Now(),
			Details: DomainTransportErrorEventDetails{
				FailureReason: PIFailureReason{
					Code:    TransportMishap,
					Message: event.FailureReason.Message,
				},
				PaymentProviderID: id,
			},
		}
	case NoSourceAccount:
		instructionEvent = PaymentInstructionEvent{
			Type:      DomainProcessingFailed,
			CreatedOn: time.Now(),
			Details: DomainRejectedEventDetails{
				FailureReason: PIFailureReason{
					Code:    NoSourceAcct,
					Message: event.FailureReason.Message,
				},
			},
		}
	case MissingFunding:
		instructionEvent = PaymentInstructionEvent{
			Type:      DomainProcessingFailed,
			CreatedOn: time.Now(),
			Details: DomainRejectedEventDetails{
				FailureReason: PIFailureReason{
					Code:    MissingFunds,
					Message: event.FailureReason.Message,
				},
			},
		}
	default:
		instructionEvent = PaymentInstructionEvent{
			Type:      DomainProcessingFailed,
			CreatedOn: time.Now(),
			Details: DomainUnhandledStatusFailedEventDetails{
				FailureReason: PIFailureReason{
					Code:    Unhandled,
					Message: event.FailureReason.Message,
				},
				PaymentProviderID: id,
			},
		}
	}

	p.events = append(p.events, instructionEvent)
}

func (p *PaymentInstruction) successfullyProcessed(paymentProviderPaymentID ProviderPaymentID) {
	p.updateStatus(Successful)
	p.events = append(p.events, PaymentInstructionEvent{
		Type:      DomainProcessingSucceeded,
		CreatedOn: time.Now(),
		Details:   DomainProcessingSucceededEventDetails{PaymentProviderPaymentID: paymentProviderPaymentID},
	})
}

func (p *PaymentInstruction) SetEvents(events []PaymentInstructionEvent) {
	p.events = events
}

func (p *PaymentInstruction) AddEvent(event PaymentInstructionEvent) {
	p.events = append(p.events, event)
}

type InvalidPaymentInstructionError struct {
	UnderlyingError error
	InstructionJSON string
}

func (i InvalidPaymentInstructionError) Error() string {
	return fmt.Sprintf("Payment Instruction Invalid, err: %v for instruction JSON: %s", i.UnderlyingError, i.InstructionJSON)
}
