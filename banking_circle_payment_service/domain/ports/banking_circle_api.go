//go:generate moq -out mocks/banking_circle_api_moq.go -pkg mocks . BankingCircleAPI

package ports

import (
	"context"

	models2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models"
	spe "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

type PaymentStatus string

// all available statuses from banking circle
// when calling endpoints [GET]/api/v1/payments/singles/{payment-id}
// and [POST] /api/v1/payments/singles
// SPS prefix are our defined values.
const (
	Processed         PaymentStatus = "Processed"         // both
	PendingProcessing PaymentStatus = "PendingProcessing" // both
	Rejected          PaymentStatus = "Rejected"          // payments/singles/{payment-id}
	MissingFunding    PaymentStatus = "MissingFunding"    // both
	// SPSUnknownValue   PaymentStatus = ""
	// ScaPending                  PaymentStatus = "ScaPending"                  // both
	// PendingApproval             PaymentStatus = "PendingApproval"             // both
	// Approved                    PaymentStatus = "Approved"                    // both
	// Unknown                     PaymentStatus = "Unknown"                     // payments/singles/{payment-id}
	// ScaExpired                  PaymentStatus = "ScaExpired"                  // payments/singles/{payment-id}
	// ScaFailed                   PaymentStatus = "ScaFailed"                   // payments/singles/{payment-id}
	// Hold                        PaymentStatus = "Hold"                        // payments/singles/{payment-id}
	// PendingCancellation         PaymentStatus = "PendingCancellation"         // payments/singles/{payment-id}
	// PendingCancellationApproval PaymentStatus = "PendingCancellationApproval" // payments/singles/{payment-id}
	// DeclinedByApprover          PaymentStatus = "DeclinedByApprover"          // payments/singles/{payment-id}
	// Cancelled                   PaymentStatus = "Cancelled"                   // payments/singles/{payment-id}
	// Reversed                    PaymentStatus = "Reversed"                    // payments/singles/{payment-id}
	// ScaDeclined                 PaymentStatus = "ScaDeclined"                 // payments/singles/{payment-id}.
)

type BankingCirclePaymentRequester interface {
	// RequestPayment makes a BC payment request call, and returns the paymentID for checking the status later.
	RequestPayment(ctx context.Context, request spe.RequestDto, slice *[]string) (spe.ResponseDto, error)
}

type BankingCirclePaymentStatusChecker interface {
	// CheckPaymentStatus uses a paymentID to check the status of the request, and returns the status.
	// This paymentID is how the BankingCircle API internally tracks the request, and will be different
	// to the PaymentInstruction ID that the Settlements Payment system uses.
	CheckPaymentStatus(paymentID models.ProviderPaymentID) (PaymentStatus, error)
}

type BankingCircleAccountBalanceChecker interface {
	// CheckAccountBalance uses a accountId to check the current balance of the account, and returns the amount.
	// This accountID is an id from BankingCircle and not the IBAN.
	CheckAccountBalance(accountID string) (models2.AccountBalance, error)
}

type BankingCircleRejectionReport interface {
	// GetRejectionReport uses a date to request the rejection report for that date, which includes multiple
	// interesting parameters as the reason of why this transaction was rejected
	GetRejectionReport(date string) (models2.RejectionReport, error)
}

type BankingCircleAPI interface {
	BankingCirclePaymentRequester
	BankingCirclePaymentStatusChecker
	BankingCircleAccountBalanceChecker
	BankingCircleRejectionReport
}
