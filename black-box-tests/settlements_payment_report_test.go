//go:build blackbox_success
// +build blackbox_success

package black_box_tests

import (
	"context"
	"os"
	"testing"

	"github.com/matryer/is"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
)

func TestPaymentReportToday(t *testing.T) {
	ctx := context.Background()
	is := is.New(t)
	client := CreateClient(ctx, t)

	err := client.GetReportToday()
	is.NoErr(err)
}

func TestPaymentReport(t *testing.T) {
	ctx := context.Background()
	is := is.New(t)
	client := CreateClient(ctx, t)

	incomingInstruction := testhelpers.NewIncomingInstructionBuilder().Build()
	_, err := client.SendPaymentInstruction(incomingInstruction)
	is.NoErr(err)

	err = client.GetReport()
	is.NoErr(err)
}

func TestCurrenciesReport(t *testing.T) {
	ctx := context.Background()
	is := is.New(t)
	client := CreateClient(ctx, t)

	incomingInstruction := testhelpers.NewIncomingInstructionBuilder().Build()
	_, err := client.SendPaymentInstruction(incomingInstruction)
	is.NoErr(err)

	err = client.GetCurrencyReport(ctx)
	is.NoErr(err)
}

func TestGetPaymentByMid(t *testing.T) {
	ctx := context.Background()
	is := is.New(t)
	client := CreateClient(ctx, t)

	incomingInstruction := testhelpers.NewIncomingInstructionBuilder().Build()
	_, err := client.SendPaymentInstruction(incomingInstruction)
	is.NoErr(err)

	_, err = client.GetPaymentByMid(ctx, incomingInstruction.Merchant.ContractNumber)
	is.NoErr(err)
}

func CreateClient(ctx context.Context, t *testing.T) *SettlementsAPIClient {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://settlements-payments.dev.saltpay.co"
	}

	client := NewClient(baseURL, os.Getenv("TEST_BEARER_TOKEN"))
	zapctx.Info(ctx, "Checking if API is up")

	err := client.CheckIfHealthy()
	if err != nil {
		zapctx.Fatal(ctx, "", zap.Error(err))
	}

	zapctx.Info(ctx, "API is up and healthy! On with the tests")

	return client
}
