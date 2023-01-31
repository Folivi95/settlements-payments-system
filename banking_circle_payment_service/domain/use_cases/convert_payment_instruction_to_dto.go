package use_cases

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/pkg/errors"

	spe "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

const maxLength = 35

var (
	// validCharactersRegex is used to sanitize data before sending to Banking Circle.
	validCharactersRegex = regexp.MustCompile(`[^a-zA-Z0-9/–?:().,‘+ &]`)
	// sqlInjectionRegexp is used to sanitize data for common sql injection patterns before sending to Banking Circle.
	sqlInjectionRegexp = regexp.MustCompile(`(?i)( (and|&|union) )|&`)
)

func ConvertPaymentInstructionToDto(ctx context.Context, paymentInstruction *models.PaymentInstruction, observer observer) (spe.RequestDto, error) {
	reqDto := spe.RequestDto{}
	reqDto.RequestedExecutiondate = paymentInstruction.IncomingInstruction.Payment.ExecutionDate
	reqDto.DebtorAccount = spe.DebtorAccount{
		Account:              paymentInstruction.IncomingInstruction.Payment.Sender.AccountNumber,
		FinancialInstitution: paymentInstruction.IncomingInstruction.Payment.Sender.BranchCode,
		Country:              "",
	}
	reqDto.DebtorViban = ""
	debtorRef := fmt.Sprintf("Settlm %s %s", paymentInstruction.IncomingInstruction.Merchant.ContractNumber, flattenDate(paymentInstruction.IncomingInstruction.Payment.ExecutionDate))
	reqDto.DebtorReference = cleanInput(ctx, debtorRef, observer, "DebtorReference")
	reqDto.DebtorNarrativeToSelf = paymentInstruction.IncomingInstruction.Merchant.ContractNumber
	reqDto.CurrencyOfTransfer = string(paymentInstruction.IncomingInstruction.IsoCode())

	amount, err := strconv.ParseFloat(paymentInstruction.IncomingInstruction.Payment.Amount, 64)
	if err != nil {
		return spe.RequestDto{}, errors.Wrap(err, "failed to parse amount as float64")
	}
	reqDto.Amount = spe.Amount{
		Currency: string(paymentInstruction.IncomingInstruction.IsoCode()),
		Amount:   amount,
	}
	reqDto.ChargeBearer = "SHA"
	reqDto.RemittanceInformation = spe.RemittanceInformation{
		Line1: cleanInput(ctx, paymentInstruction.IncomingInstruction.Merchant.ContractNumber, observer, "RemittanceInformation.Line1"),
		Line2: cleanInput(ctx, "Paydate "+flattenDate(paymentInstruction.IncomingInstruction.Payment.ExecutionDate), observer, "RemittanceInformation.Line2"),
		Line3: cleanInput(ctx, "Settlm", observer, "RemittanceInformation.Line3"),
		Line4: "",
	}
	reqDto.CreditorID = ""
	reqDto.CreditorAccount = spe.CreditorAccount{
		Account:              paymentInstruction.IncomingInstruction.Merchant.Account.AccountNumber,
		FinancialInstitution: paymentInstruction.IncomingInstruction.Merchant.Account.Swift,
		Country:              paymentInstruction.IncomingInstruction.Merchant.Account.Country,
	}

	reqDto.CreditorName = cleanInput(ctx, paymentInstruction.IncomingInstruction.Merchant.Name, observer, "CreditorName")
	reqDto.CreditorAddress = spe.CreditorAddress{
		Line1: cleanInput(ctx, paymentInstruction.IncomingInstruction.Merchant.Address.AddressLine1, observer, "CreditorAddress.Line1"),
		Line2: cleanInput(ctx, makeCreditorAddressLine2(paymentInstruction.IncomingInstruction.Merchant.Address), observer, "CreditorAddress.Line2"),
		Line3: cleanInput(ctx, makeCreditorAddressLine3(paymentInstruction.IncomingInstruction.Merchant.Address), observer, "CreditorAddress.Line3"),
	}

	return reqDto, nil
}

func makeCreditorAddressLine2(address models.Address) string {
	// if line2 exists on the source, use it, or create as a combo of "country city"
	if address.AddressLine2 != "" {
		return address.AddressLine2
	}
	return countryCityCombo(address)
}

func makeCreditorAddressLine3(address models.Address) string {
	// if line2 does not exist on the source, line 2 was already made to be country-city combo; no need of line 3
	if address.AddressLine2 == "" {
		return ""
	}
	return countryCityCombo(address)
}

func countryCityCombo(address models.Address) string {
	return address.Country + " " + address.City
}

// flattenDate takes a time object and returns a date of format yyyymmdd.
func flattenDate(date time.Time) string {
	return date.Format("20060102")
}

func cleanInput(ctx context.Context, inputValue string, observer observer, fieldName string) string {
	newInput := inputValue

	newInput = validCharactersRegex.ReplaceAllString(newInput, ``)
	newInput = sqlInjectionRegexp.ReplaceAllString(newInput, " ")

	// truncate to 35 characters
	n := maxLength
	if len(newInput) > maxLength {
		for !utf8.ValidString(newInput[:n]) {
			n--
		}
		newInput = newInput[:n]
	}

	if newInput != inputValue {
		observer.CleanInput(ctx, inputValue, newInput, fieldName)
	}

	return newInput
}
