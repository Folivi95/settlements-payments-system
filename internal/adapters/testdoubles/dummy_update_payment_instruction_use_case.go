package testdoubles

import (
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type DummyUpdatePaymentInstructionUseCase struct{}

func (d DummyUpdatePaymentInstructionUseCase) Execute(models.PaymentProviderEvent) error {
	return nil
}
