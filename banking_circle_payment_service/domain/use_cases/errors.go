package use_cases

import (
	"fmt"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type TransportError struct {
	UnderlyingError error
	ID              models.PaymentInstructionID
	ContractNumber  string
}

func (t TransportError) Error() string {
	return fmt.Sprintf("http call instruction payment failed for payment instruction %s (merchant contract number %s), error: %v", t.ID, t.ContractNumber, t.UnderlyingError)
}

type NoSourceAccountError struct {
	IsoCode  string
	HighRisk bool
}

func (n NoSourceAccountError) Error() string {
	return fmt.Sprintf("account Not Found for currency %s and highrisk as %t", n.IsoCode, n.HighRisk)
}

type BankingCircleError struct {
	Status ports.PaymentStatus
}

func (b BankingCircleError) Error() string {
	return fmt.Sprintf("Payment provider status %+v", b.Status)
}

type InvalidAccountIDError struct {
	AccountID string
}

func (i InvalidAccountIDError) Error() string {
	return fmt.Sprintf("No valid account for the provided account id: %s", i.AccountID)
}
