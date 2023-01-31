//go:build blackbox_failure
// +build blackbox_failure

package black_box_tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
)

func TestSettlementsSystem(t *testing.T) {
	client := NewClient(os.Getenv("BASE_URL"), os.Getenv("TEST_BEARER_TOKEN"))

	err := client.CheckIfHealthy()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("rejected payment by bc recorded", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		is := is.New(t)

		rejectedMerchantContractNumber := "rejected-czk-test-merchant"

		payment := models.Payment{
			Sender: models.Sender{},
			Amount: "100",
			Currency: models.Currency{
				IsoCode:   "CZK",
				IsoNumber: "203",
			},
			ExecutionDate: time.Now().UTC(),
		}

		incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithPayment(payment).WithMerchantContractNumber(rejectedMerchantContractNumber).Build()

		id, err := client.SendPaymentInstruction(incomingInstruction)
		is.NoErr(err)

		is.NoErr(client.CheckFailedPayment(ctx, id))
	})

	t.Run("generate rejection report", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		is := is.New(t)

		rejectedMerchantContractNumber := "rejected-czk-test-merchant"

		incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithMerchantContractNumber(rejectedMerchantContractNumber).Build()

		_, err = client.SendPaymentInstruction(incomingInstruction)
		is.NoErr(err)

		time.Sleep(5 * time.Second)
		is.NoErr(client.RejectionReport(ctx))
	})

	t.Run("invalid payment currency (checking stuff before bc gets recorded)", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		is := is.New(t)

		badCurrency := models.Currency{
			IsoCode:   "wfwgwef",
			IsoNumber: "wewfwwwwww",
		}
		incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithCurrency(badCurrency).Build()

		id, err := client.SendPaymentInstruction(incomingInstruction)
		is.NoErr(err)

		is.NoErr(client.CheckFailedPayment(ctx, id))
	})

	t.Run("We have a DLQ endpoint", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		time.Sleep(5 * time.Second) // todo Implement retries rather than sleeping.
		is.NoErr(client.CheckUnprocessedPayments())
	})
}
