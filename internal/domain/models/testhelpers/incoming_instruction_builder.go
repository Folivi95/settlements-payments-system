package testhelpers

import (
	"fmt"
	"time"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

type IncomingInstructionBuilder struct {
	IncomingInstruction models.IncomingInstruction
}

func NewIncomingInstructionBuilder() *IncomingInstructionBuilder {
	now := time.Now()
	amount := fmt.Sprintf("%d%d%d%d%d",
		now.Hour()+1,
		now.Minute(),
		now.Second(),
		now.Nanosecond()/1000000,
		testhelpers.RandomIntWithin(1, 1000),
	)
	return &IncomingInstructionBuilder{IncomingInstruction: models.IncomingInstruction{
		Merchant: models.Merchant{
			ContractNumber: "1234567",
			RegNumber:      "1234567890",
			Name:           testhelpers.RandomString(),
			Email:          testhelpers.RandomString(),
			Address: models.Address{
				AddressLine1: testhelpers.RandomString(),
			},
			Account: models.Account{
				AccountNumber:        "DE1111111111111111",
				Swift:                "567",
				Country:              "DE",
				SwiftReferenceNumber: "999",
			},
			HighRisk: false,
		},
		Payment: models.Payment{
			Sender: models.Sender{},
			Amount: amount,
			Currency: models.Currency{
				IsoCode:   "EUR",
				IsoNumber: "978",
			},
			ExecutionDate: time.Now().UTC().Round(time.Microsecond),
		},
		Metadata: models.Metadata{
			Source:   "way4",
			Filename: testhelpers.RandomString(),
			FileType: "ufx",
			Sender:   "SAXO",
		},
	}}
}

func (i *IncomingInstructionBuilder) WithHighRIsk() *IncomingInstructionBuilder {
	i.IncomingInstruction.Merchant.HighRisk = true
	return i
}

func (i *IncomingInstructionBuilder) WithMerchantName(name string) *IncomingInstructionBuilder {
	i.IncomingInstruction.Merchant.Name = name
	return i
}

func (i *IncomingInstructionBuilder) WithPayment(paymentInformation models.Payment) *IncomingInstructionBuilder {
	i.IncomingInstruction.Payment = paymentInformation
	return i
}

func (i *IncomingInstructionBuilder) WithMerchantAccountNumber(number string) *IncomingInstructionBuilder {
	i.IncomingInstruction.Merchant.Account.AccountNumber = number
	return i
}

func (i *IncomingInstructionBuilder) WithMerchantAccountCountry(code string) *IncomingInstructionBuilder {
	i.IncomingInstruction.Merchant.Account.Country = code
	return i
}

func (i *IncomingInstructionBuilder) WithCurrency(currency models.Currency) *IncomingInstructionBuilder {
	i.IncomingInstruction.Payment.Currency = currency
	return i
}

func (i *IncomingInstructionBuilder) WithAddressLine1(addressLine string) *IncomingInstructionBuilder {
	i.IncomingInstruction.Merchant.Address.AddressLine1 = addressLine
	return i
}

func (i *IncomingInstructionBuilder) WithPaymentExecutionDate(date time.Time) *IncomingInstructionBuilder {
	i.IncomingInstruction.Payment.ExecutionDate = date
	return i
}

func (i *IncomingInstructionBuilder) WithPaymentAmount(amount string) *IncomingInstructionBuilder {
	i.IncomingInstruction.Payment.Amount = amount
	return i
}

func (i *IncomingInstructionBuilder) WithPaymentCurrencyISONumber(isoNumber string) *IncomingInstructionBuilder {
	i.IncomingInstruction.Payment.Currency.IsoNumber = isoNumber
	return i
}

func (i *IncomingInstructionBuilder) WithPaymentCurrencyISOCode(isoCode models.CurrencyCode) *IncomingInstructionBuilder {
	i.IncomingInstruction.Payment.Currency.IsoCode = isoCode
	return i
}

func (i *IncomingInstructionBuilder) WithMetadataSource(metadataSource string) *IncomingInstructionBuilder {
	i.IncomingInstruction.Metadata.Source = metadataSource
	return i
}

func (i *IncomingInstructionBuilder) WithMetadataSender(metadataSender string) *IncomingInstructionBuilder {
	i.IncomingInstruction.Metadata.Sender = metadataSender
	return i
}

func (i *IncomingInstructionBuilder) WithMerchantContractNumber(number string) *IncomingInstructionBuilder {
	i.IncomingInstruction.Merchant.ContractNumber = number
	return i
}

func (i *IncomingInstructionBuilder) WithMerchantRegNumber(number string) *IncomingInstructionBuilder {
	i.IncomingInstruction.Merchant.RegNumber = number
	return i
}

func (i *IncomingInstructionBuilder) WithCorrelationID(corrID string) *IncomingInstructionBuilder {
	i.IncomingInstruction.PaymentCorrelationId = corrID
	return i
}

func (i *IncomingInstructionBuilder) WithMerchantAccountSwift(swift string) *IncomingInstructionBuilder {
	i.IncomingInstruction.Merchant.Account.Swift = swift
	return i
}

func (i *IncomingInstructionBuilder) Build() models.IncomingInstruction {
	return i.IncomingInstruction
}
