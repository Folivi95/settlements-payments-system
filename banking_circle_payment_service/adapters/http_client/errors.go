package http_client

import (
	"fmt"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
)

type (
	BankingCircleActionType string
	UniqueID                string
)

const (
	RequestPayment  BankingCircleActionType = "making payment"
	CheckingPayment BankingCircleActionType = "checking payment"
	RejectionReport BankingCircleActionType = "rejection report"
)

type UnrecognisedBankingCircleError struct {
	URL      string
	Status   int
	Body     string
	UniqueID UniqueID
	Action   BankingCircleActionType
}

func (u UnrecognisedBankingCircleError) Error() string {
	return fmt.Sprintf("error code from server when %s, %d, %s with requestID %s", u.Action, u.Status, u.Body, u.UniqueID)
}

type UnauthorisedWithBankingCircleError struct {
	URL          string
	ResponseBody string
	UniqueID     UniqueID
}

func (u UnauthorisedWithBankingCircleError) Error() string {
	return fmt.Sprintf("unauthorised to make call with banking circle at: %s, response: %s, and requestID %s", u.URL, u.ResponseBody, u.UniqueID)
}

type PaymentNotFoundError struct {
	URL       string
	PaymentID string
	UniqueID  UniqueID
}

func (p PaymentNotFoundError) Error() string {
	return fmt.Sprintf("banking circle could not find payment %s from %s with requestID %s", p.PaymentID, p.URL, p.UniqueID)
}

type InvalidPaymentRequestError struct {
	Request      single_payment_endpoint.RequestDto
	ErrorMessage string
	UniqueID     UniqueID
}

func (i InvalidPaymentRequestError) Error() string {
	return fmt.Sprintf("banking circle reports payment with requestID %s as invalid: %s, %+v", i.UniqueID, i.ErrorMessage, i.Request)
}
