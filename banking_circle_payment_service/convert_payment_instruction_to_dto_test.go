//go:build unit
// +build unit

package banking_circle_payment_service_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	is2 "github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/use_cases"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
)

func TestConvertPaymentInstructionToDTO(t *testing.T) {
	var (
		ctx = context.Background()
		is  = is2.New(t)
	)

	dummyObserver := use_cases.NewObserver(testdoubles.DummyMetricsClient{})

	type creditorNameTest struct {
		creditorName string
		badText      string
		wantedName   string
	}

	creditorNameTestCases := []creditorNameTest{
		{creditorName: "Catarina&Riya", badText: "&", wantedName: "Catarina Riya"},
		{creditorName: "Catarina & Riya", badText: " & ", wantedName: "Catarina Riya"},
		{creditorName: "Catarina AnD Riya", badText: " and ", wantedName: "Catarina Riya"},
		{creditorName: "Catarina uNiOn Riya", badText: " union ", wantedName: "Catarina Riya"},
		{creditorName: "Catarina \"Riya", badText: "\"", wantedName: "Catarina Riya"},
	}

	for _, tc := range creditorNameTestCases {
		t.Run(fmt.Sprintf("removes %q from the creditor name field", tc.badText), func(t *testing.T) {
			paymentInstruction := models.PaymentInstruction{IncomingInstruction: testhelpers.NewIncomingInstructionBuilder().WithMerchantName(tc.creditorName).Build()}

			dto, err := use_cases.ConvertPaymentInstructionToDto(ctx, &paymentInstruction, dummyObserver)

			is.NoErr(err)
			is.Equal(dto.CreditorName, tc.wantedName)
		})
	}

	t.Run("truncate creditor name if longer than 35 characters", func(t *testing.T) {
		name := strings.Repeat("a", 34) + "é" // it should handle characters with accents
		expectedName := strings.Repeat("a", 34)

		paymentInstruction := models.PaymentInstruction{IncomingInstruction: testhelpers.NewIncomingInstructionBuilder().WithMerchantName(name).Build()}

		dto, err := use_cases.ConvertPaymentInstructionToDto(ctx, &paymentInstruction, dummyObserver)

		is.NoErr(err)
		is.Equal(dto.CreditorName, expectedName)
	})
}
