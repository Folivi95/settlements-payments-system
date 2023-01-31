//go:build blackbox_success
// +build blackbox_success

package black_box_tests

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/saltpay/go-kafka-driver"
	zapctx "github.com/saltpay/go-zap-ctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"
	testhelpers2 "github.com/saltpay/settlements-payments-system/internal/testhelpers"
)

// This test will treat the Payments Service as a black box.

func TestTrackSuccessfulPayments(t *testing.T) {
	// If I put a valid payment into the system
	// Wait a while
	// Then check the status of the payments
	// Then I should see that it WAS successful
	// And the successful payment details should match what I put in
	var (
		ctx = context.Background()
		err error

		baseURL                  = os.Getenv("BASE_URL")
		kafkaEndpoint            = os.Getenv("KAFKA_ENDPOINT")
		kafkaUsername            = os.Getenv("KAFKA_USERNAME")
		kafkaPassword            = os.Getenv("KAFKA_PASSWORD")
		transactionsTopic        = os.Getenv("KAFKA_TOPICS_TRANSACTIONS")
		paymentStateUpdatesTopic = os.Getenv("KAFKA_TOPICS_PAYMENT_STATE_UPDATES")
		mockISBServiceStr        = os.Getenv("MOCK_ISB_SERVICE")
	)

	if baseURL == "" {
		baseURL = "https://settlements-payments.dev.saltpay.co"
	}

	require.NotEmpty(t, kafkaEndpoint, "kafka endpoint undefined")
	require.NotEmpty(t, transactionsTopic, "transactions topic undefined")
	require.NotEmpty(t, paymentStateUpdatesTopic, "Payment State Updates Topic topic undefined")

	// TODO: black box test locally is a service test, but acts as system test in the post deployment hook and in the blackBoxTests
	// We need to convert it to a normal service test as what we have in ISB
	// This MOCK_ISB_SERVICE is just a temp solution
	shouldMockISBService := false
	if mockISBServiceStr != "" {
		shouldMockISBService, err = strconv.ParseBool(mockISBServiceStr)
		require.NoError(t, err, "MOCK_ISB_SERVICE env variable should be a bool")
	}

	client := NewClient(baseURL, os.Getenv("TEST_BEARER_TOKEN"))

	zapctx.Info(ctx, "Checking if API is up")
	err = client.CheckIfHealthy()
	require.NoError(t, err)

	zapctx.Info(ctx, "API is up and healthy! On with the tests")

	t.Run("BC payments", func(t *testing.T) {
		t.Run("kafka client test", func(t *testing.T) {
			t.Parallel()
			if v := os.Getenv("KAFKA_TEST_ENABLED"); v == "" {
				t.Skip("kafka black box tests disabled")
			}
			var (
				ctx                 = context.Background()
				incomingInstruction = testhelpers.NewIncomingInstructionBuilder().Build()
				mid                 = fmt.Sprintf("kafkaTest-%s", uuid.New().String())
			)

			incomingInstruction.Merchant.ContractNumber = mid
			v, err := json.Marshal(incomingInstruction)
			require.NoError(t, err)

			producer, err := kafka.NewProducer(ctx, kafka.ProducerConfig{
				Addr:     strings.Split(kafkaEndpoint, ","),
				Topic:    transactionsTopic,
				Username: kafkaUsername,
				Password: kafkaPassword,
			})
			require.NoError(t, err)
			err = producer.WriteMessage(ctx, kafka.Message{
				Value: v,
			})
			require.NoError(t, err)
			producer.Close()

			// TODO: Check if we should continue to poll for a success / terminal state
			payment, err := client.WaitUntilPaymentIsInState(ctx, *client, mid, models.Successful)
			assert.NoError(t, err)
			assert.Equal(t, models.Successful, payment.GetStatus())
		})

		t.Run("http client test", func(t *testing.T) {
			t.Parallel()
			incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithMerchantContractNumber(fmt.Sprintf("httpTest-%s", uuid.New().String())).Build()
			id, err := client.SendPaymentInstruction(incomingInstruction)
			require.NoError(t, err)

			l := fmt.Sprintf("Payment instruction with ID %s sent. Now checking for success status...", id)
			fmt.Println(l)

			assert.NoError(t, client.CheckPaymentWasSuccessful(ctx, id))
		})

		t.Run("S3 bucket test", func(t *testing.T) {
			t.Parallel()

			var (
				mid = fmt.Sprintf("s3Test-%s", uuid.New().String())
				ufx = testhelpers.NewUfxBuilder().WithMerchantID(mid).Build()
			)

			err := client.PutUfxFileIntoS3Bucket(ufx)
			require.NoError(t, err)

			payment, err := client.WaitUntilPaymentIsInState(ctx, *client, mid, models.Successful)
			assert.NoError(t, err)
			assert.Equal(t, models.Successful, payment.GetStatus())
		})
	})

	t.Run("ISB payments", func(t *testing.T) {
		// Given an ufx file with 1 isk payment
		t.Parallel()
		var (
			ctx = context.Background()
			mid = fmt.Sprintf("isbTest-%s", testhelpers2.RandomString())
			ufx = testhelpers.NewUfxBuilder().
				WithSender("RB").
				WithCurrency("ISK").
				WithIBAN("IS090123261234561234567890").
				WithMerchantID(mid).
				Build()
		)

		// When ufx file is sent to the S3 bucket
		err := client.PutUfxFileIntoS3Bucket(ufx)
		require.NoError(t, err)

		if shouldMockISBService {
			// Then payment should get Submitted to the payment provider
			payment, err := client.WaitUntilPaymentIsInState(ctx, *client, mid, models.SubmittedForProcessing)
			require.NoError(t, err)

			// And given a state update message for this payment
			producer, err := kafka.NewProducer(ctx, kafka.ProducerConfig{
				Addr:     strings.Split(kafkaEndpoint, ","),
				Topic:    paymentStateUpdatesTopic,
				Username: kafkaUsername,
				Password: kafkaPassword,
			})
			require.NoError(t, err)

			// When the state update message is sent into the kafka topic
			err = producer.WriteMessage(ctx, kafka.Message{
				Value: []byte(fmt.Sprintf("{\"payment_instruction_id\":\"%s\", \"updated_state\":\"PROCESSED\"}", payment.ID())),
			})
			require.NoError(t, err)
			producer.Close()
		}

		// Then payment state should be Successful
		_, err = client.WaitUntilPaymentIsInState(ctx, *client, mid, models.Successful)
		require.NoError(t, err)
	})
}
