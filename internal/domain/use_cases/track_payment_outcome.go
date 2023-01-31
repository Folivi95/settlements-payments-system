package use_cases

import (
	"context"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
	"github.com/saltpay/settlements-payments-system/internal/domain/validation"
)

type TrackPaymentOutcome struct {
	paymentInstructionRepo  ports.StorePaymentInstructionToRepo
	paymentExporterProducer ports.PaymentExporterProducer
	eventValidator          validation.PPEventValidator
}

func NewTrackPaymentOutcome(
	repo ports.StorePaymentInstructionToRepo,
	eventValidator validation.PPEventValidator,
	acquiringHostProducer ports.PaymentExporterProducer,
) TrackPaymentOutcome {
	return TrackPaymentOutcome{
		paymentInstructionRepo:  repo,
		eventValidator:          eventValidator,
		paymentExporterProducer: acquiringHostProducer,
	}
}

func (u TrackPaymentOutcome) Execute(ctx context.Context, ppEvent models.PaymentProviderEvent) error {
	if err := u.eventValidator.Validate(ppEvent); err != nil {
		return err
	}

	pi := &ppEvent.PaymentInstruction
	pi.TrackPPEvent(ppEvent)

	lastEvent := pi.Events()[len(pi.Events())-1]

	err := u.paymentInstructionRepo.UpdatePayment(ctx, pi.ID(), pi.GetStatus(), lastEvent)
	if err != nil {
		return err
	}

	return u.paymentExporterProducer.ReportPaymentStatus(ctx, ppEvent)
}
