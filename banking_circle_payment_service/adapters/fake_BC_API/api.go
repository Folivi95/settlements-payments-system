package fakebcapi

import (
	"context"
	"time"

	models2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	spe "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

const (
	accountNumberThatWillGetRejected = "987"
)

type FakeBankingCircleAPI struct{}

func NewFakeBankingCircleAPI() *FakeBankingCircleAPI {
	return &FakeBankingCircleAPI{}
}

func (f *FakeBankingCircleAPI) RequestPayment(_ context.Context, request spe.RequestDto, _ *[]string) (spe.ResponseDto, error) {
	switch request.Amount.Amount {
	case 100:
		return spe.ResponseDto{
			PaymentID: accountNumberThatWillGetRejected,
			Status:    string(ports.PendingProcessing),
		}, nil
	default:
		return spe.ResponseDto{
			PaymentID: "123",
			Status:    string(ports.Processed),
		}, nil
	}
}

func (f *FakeBankingCircleAPI) CheckPaymentStatus(paymentInstructionID models.ProviderPaymentID) (ports.PaymentStatus, error) {
	switch paymentInstructionID {
	case accountNumberThatWillGetRejected:
		return ports.Rejected, nil
	default:
		return ports.Processed, nil
	}
}

func (f *FakeBankingCircleAPI) GetRejectionReport(_ string) (models2.RejectionReport, error) {
	return models2.RejectionReport{Rejections: []models2.Rejection{{
		PTxndate:   time.Now(),
		ReportDate: time.Now(),
		ValueDate:  time.Now(),
	}}}, nil
}

func (f *FakeBankingCircleAPI) CheckAccountBalance(_ string) (models2.AccountBalance, error) {
	return models2.AccountBalance{
		Result: []models2.Balance{
			{
				Currency:         models.EUR,
				BeginOfDayAmount: 100_000_000,
			},
		},
	}, nil
}
