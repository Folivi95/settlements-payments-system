package testhelpers

import (
	"time"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type PaymentInstructionBuilder struct {
	PaymentInstruction models.PaymentInstruction
}

func NewPaymentInstructionBuilder() *PaymentInstructionBuilder {
	return &PaymentInstructionBuilder{PaymentInstruction: models.NewPaymentInstruction(NewIncomingInstructionBuilder().Build())}
}

func (p *PaymentInstructionBuilder) WithIncomingInstruction(incomingInstruction models.IncomingInstruction) *PaymentInstructionBuilder {
	p.PaymentInstruction.IncomingInstruction = incomingInstruction
	return p
}

func (p *PaymentInstructionBuilder) WithPaymentProvider(paymentProvider models.PaymentProviderType) *PaymentInstructionBuilder {
	p.PaymentInstruction.SetPaymentProvider(paymentProvider)
	return p
}

func (p *PaymentInstructionBuilder) WithStatus(status models.PaymentInstructionStatus) *PaymentInstructionBuilder {
	p.PaymentInstruction.SetStatus(status)
	return p
}

func (p *PaymentInstructionBuilder) WithMid(mid string) *PaymentInstructionBuilder {
	p.PaymentInstruction.IncomingInstruction.Merchant.ContractNumber = mid
	return p
}

func (p *PaymentInstructionBuilder) WithCurrency(currency models.Currency) *PaymentInstructionBuilder {
	p.PaymentInstruction.IncomingInstruction.Payment.Currency = currency
	return p
}

func (p *PaymentInstructionBuilder) WithAmount(amount string) *PaymentInstructionBuilder {
	p.PaymentInstruction.IncomingInstruction.Payment.Amount = amount
	return p
}

func (p *PaymentInstructionBuilder) WithAccountNumber(accountNumber string) *PaymentInstructionBuilder {
	p.PaymentInstruction.IncomingInstruction.Merchant.Account.AccountNumber = accountNumber
	return p
}

func (p *PaymentInstructionBuilder) WithDate(date time.Time) *PaymentInstructionBuilder {
	p.PaymentInstruction.IncomingInstruction.Payment.ExecutionDate = date
	return p
}

func (p *PaymentInstructionBuilder) WithEvents(events []models.PaymentInstructionEvent) *PaymentInstructionBuilder {
	p.PaymentInstruction.SetEvents(events)
	return p
}

func (p *PaymentInstructionBuilder) Build() models.PaymentInstruction {
	return p.PaymentInstruction
}
