package dto

import (
	"encoding/json"
	"time"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type (
	IncomingInstruction struct {
		Merchant      Merchant `json:"merchant"`
		Metadata      Metadata `json:"metadata"`
		Payment       Payment  `json:"payment"`
		CorrelationID string   `json:"correlationId"`
	}

	Address struct {
		Country      string `json:"country"`
		City         string `json:"city"`
		AddressLine1 string `json:"addressLine1"`
		AddressLine2 string `json:"addressLine2"`
	}

	Account struct {
		AccountNumber        string `json:"accountNumber"`
		Swift                string `json:"swift"`
		Country              string `json:"country"`
		SwiftReferenceNumber string `json:"swiftReferenceNumber"`
	}

	Merchant struct {
		ContractNumber string  `json:"contractNumber"`
		Name           string  `json:"name"`
		Email          string  `json:"email"`
		Address        Address `json:"address"`
		Account        Account `json:"account"`
		HighRisk       bool    `json:"highRisk"`
	}

	Metadata struct {
		Source   string `json:"source"`
		Filename string `json:"filename"`
		FileType string `json:"fileType"`
	}

	Currency struct {
		IsoCode   models.CurrencyCode `json:"isoCode"`
		IsoNumber string              `json:"isoNumber"`
	}

	Payment struct {
		Sender        Sender    `json:"sender"`
		Amount        string    `json:"amount"`
		Currency      Currency  `json:"currency"`
		ExecutionDate time.Time `json:"executionDate"`
	}

	Sender struct {
		Name          string `json:"name"`
		AccountNumber string `json:"accountNumber"`
		BranchCode    string `json:"branchCode"`
	}
)

func (i *IncomingInstruction) MapFromIncomingInstructionKafkaDTO() models.IncomingInstruction {
	return models.IncomingInstruction{
		Merchant:             i.Merchant.mapFromMerchantKafkaDTO(),
		Metadata:             i.Metadata.mapFromMetadataKafkaDTO(),
		Payment:              i.Payment.mapFromPaymentKafkaDTO(),
		PaymentCorrelationId: i.CorrelationID,
	}
}

func (m *Merchant) mapFromMerchantKafkaDTO() models.Merchant {
	return models.Merchant{
		ContractNumber: m.ContractNumber,
		Name:           m.Name,
		Email:          m.Email,
		Address: models.Address{
			Country:      m.Address.Country,
			City:         m.Address.City,
			AddressLine1: m.Address.AddressLine1,
			AddressLine2: m.Address.AddressLine2,
		},
		Account: models.Account{
			AccountNumber:        m.Account.AccountNumber,
			Swift:                m.Account.Swift,
			Country:              m.Account.Country,
			SwiftReferenceNumber: m.Account.SwiftReferenceNumber,
		},
		HighRisk: m.HighRisk,
	}
}

func (m *Metadata) mapFromMetadataKafkaDTO() models.Metadata {
	return models.Metadata{
		Source:   m.Source,
		Filename: m.Filename,
		FileType: m.FileType,
	}
}

func (p *Payment) mapFromPaymentKafkaDTO() models.Payment {
	return models.Payment{
		Sender: models.Sender{
			Name:          p.Sender.Name,
			AccountNumber: p.Sender.AccountNumber,
			BranchCode:    p.Sender.BranchCode,
		},
		Amount: p.Amount,
		Currency: models.Currency{
			IsoCode:   p.Currency.IsoCode,
			IsoNumber: p.Currency.IsoNumber,
		},
		ExecutionDate: p.ExecutionDate,
	}
}

func NewIncomingInstructionKafkaDTOFromBytes(in []byte) (IncomingInstruction, error) {
	var incomingInstruction IncomingInstruction
	err := json.Unmarshal(in, &incomingInstruction)
	return incomingInstruction, err
}
