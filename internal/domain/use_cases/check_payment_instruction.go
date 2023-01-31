package use_cases

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type CheckPaymentInstruction struct {
	repo ports.GetPaymentInstructionFromRepo
}

func NewCheckPaymentInstruction(repo ports.GetPaymentInstructionFromRepo) CheckPaymentInstruction {
	return CheckPaymentInstruction{repo: repo}
}

func (c CheckPaymentInstruction) Execute(ctx context.Context, id models.PaymentInstructionID) (models.PaymentInstruction, error) {
	paymentInstruction, err := c.repo.Get(ctx, id)
	if err != nil {
		return models.PaymentInstruction{}, err
	}

	return paymentInstruction, err
}

func (c CheckPaymentInstruction) RetrieveByCorrelationID(ctx context.Context, correlationID string) ([]models.PaymentInstruction, error) {
	paymentInstructions, err := c.repo.GetFromCorrelationID(ctx, correlationID)
	if err != nil {
		return paymentInstructions, err
	}

	return paymentInstructions, err
}
