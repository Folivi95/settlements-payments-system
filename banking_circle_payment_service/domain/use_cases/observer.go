package use_cases

import (
	"context"
	"time"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
	ports2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type observer struct {
	MetricsClient ports.MetricsClient
}

func NewObserver(metrics ports.MetricsClient) observer {
	return observer{
		MetricsClient: metrics,
	}
}

const (
	requestPaymentCounterName     = "app_settlements_provider_request_payment"
	checkPaymentStatusCounterName = "app_settlements_provider_check_payment_status"
	paymentProviderTag            = "banking_circle"
	successTag                    = "success"
	failedTag                     = "failure"
	failedUnauthorizedTag         = "failure_unauthorized"
	failedBadRequestTag           = "failure_bad_request"
	missingFunds                  = "app_settlements_provider_missing_funds"
	dataFieldCleaned              = "app_settlements_provider_data_cleaned"
	successfulPaymentCounterName  = "app_settlements_provider_success_payment"
)

var (
	successTags            = []string{successTag, paymentProviderTag}
	failedTags             = []string{failedTag, paymentProviderTag}
	failedBadRequestTags   = []string{failedBadRequestTag, paymentProviderTag}
	failedUnauthorizedTags = []string{failedUnauthorizedTag, paymentProviderTag}
)

func (m observer) RequestPaymentSucceeded(ctx context.Context, instruction models.PaymentInstruction, _ single_payment_endpoint.ResponseDto) {
	m.MetricsClient.Count(ctx, requestPaymentCounterName, 1, append(successTags, string(instruction.IncomingInstruction.IsoCode())))
}

func (m observer) RequestPaymentFailed(ctx context.Context, instruction models.PaymentInstruction, slice []string, err error) {
	m.MetricsClient.Count(ctx, requestPaymentCounterName, 1, append(failedTags, string(instruction.IncomingInstruction.IsoCode())))
	zapctx.Error(ctx, "Banking Circle request payment call for payment instruction failed",
		zap.String("id", string(instruction.ID())),
		zap.String("merchant_contract_number", string(instruction.ID())),
		zap.Strings("uuids", slice),
		zap.Error(err),
	)
}

func (m observer) RequestPaymentFailedUnauthorized(ctx context.Context, instruction models.PaymentInstruction, slice []string, err error) {
	m.MetricsClient.Count(ctx, requestPaymentCounterName, 1, append(failedUnauthorizedTags, string(instruction.IncomingInstruction.IsoCode())))
	zapctx.Error(ctx, "Banking Circle request payment call for payment instruction failed",
		zap.String("id", string(instruction.ID())),
		zap.String("merchant_contract_number", string(instruction.ID())),
		zap.Strings("uuids", slice),
		zap.Error(err),
	)
}

func (m observer) RequestPaymentFailedBadRequest(ctx context.Context, instruction models.PaymentInstruction, slice []string, err error) {
	m.MetricsClient.Count(ctx, requestPaymentCounterName, 1, append(failedBadRequestTags, string(instruction.IncomingInstruction.IsoCode())))
	zapctx.Error(ctx, "Banking Circle request payment call for payment instruction failed",
		zap.String("id", string(instruction.ID())),
		zap.String("merchant_contract_number", string(instruction.ID())),
		zap.Strings("uuids", slice),
		zap.Error(err),
	)
}

func (m observer) CheckPaymentSucceeded(ctx context.Context, instruction models.PaymentInstruction, status ports2.PaymentStatus) {
	m.MetricsClient.Count(ctx, checkPaymentStatusCounterName, 1, append(successTags, string(instruction.IncomingInstruction.IsoCode())))
	zapctx.Debug(ctx, "flow_step #10a: checked status in Banking Circle for payment instruction",
		zap.String("id", string(instruction.ID())),
		zap.String("status", string(status)),
	)
}

func (m observer) CheckPaymentFailed(ctx context.Context, instruction models.PaymentInstruction, err error) {
	m.MetricsClient.Count(ctx, checkPaymentStatusCounterName, 1, append(failedTags, string(instruction.IncomingInstruction.IsoCode())))
	zapctx.Info(ctx, "flow_step #10b: ERROR when checking status in Banking Circle for payment instruction. Retrying...",
		zap.String("id", string(instruction.ID())),
		zap.String("contract_number", instruction.ContractNumber()),
		zap.Error(err),
	)
}

func (m observer) PaymentIsUnprocessed(ctx context.Context, id models.PaymentInstructionID, mid string, status models.PPEventFailureCode) {
	zapctx.Error(ctx, "[CheckBankingCirclePaymentStatus] (Execute) payment for instruction has failed processing",
		zap.String("id", string(id)),
		zap.String("merchant_contract_number", mid),
		zap.String("status", string(status)),
	)
}

func (m observer) StillPending(ctx context.Context, id models.PaymentInstructionID, mid string, attempts int) {
	zapctx.Debug(ctx, "[CheckBankingCirclePaymentStatus] (loopCheckPaymentStatus) payment instruction status is still pending",
		zap.String("id", string(id)),
		zap.String("merchant_contract_number", mid),
		zap.Int("iteration", attempts),
	)
}

func (m observer) NoLongerPending(ctx context.Context, id models.PaymentInstructionID, mid string, status ports2.PaymentStatus, attempts int) {
	zapctx.Debug(ctx, "[CheckBankingCirclePaymentStatus] (loopCheckPaymentStatus) status is no longer pending for payment instruction",
		zap.String("id", string(id)),
		zap.String("merchant_contract_number", mid),
		zap.String("status", string(status)),
		zap.Int("iteration", attempts),
	)
}

func (m observer) RequestedPayment(ctx context.Context, instructionID models.PaymentInstructionID, paymentID models.ProviderPaymentID) {
	zapctx.Debug(ctx, "flow_step #9: payment instruction sent to Banking Circle API",
		zap.String("id", string(instructionID)),
		zap.String("banking_circle_id", string(paymentID)),
	)
}

func (m observer) CheckPaymentStatusTotallyFailed(ctx context.Context, id models.PaymentInstructionID, err error, retries int) {
	zapctx.Error(ctx, "[CheckBankingCirclePaymentStatus] (Execute) Banking Circle check payment status call totally failed for payment instruction",
		zap.String("id", string(id)),
		zap.Int("retries", retries),
		zap.Error(err),
	)
}

func (m observer) FinishedProcessing(
	ctx context.Context,
	paymentInstruction models.PaymentInstruction,
	processingSucceeded bool,
	bcProcessingStatus ports2.PaymentStatus,
	processingDuration time.Duration,
) {
	if processingSucceeded {
		zapctx.Debug(ctx, "flow_step #11a: Banking Circle payment processed successfully for payment instruction",
			zap.String("id", string(paymentInstruction.ID())),
			zap.String("merchant_contract_number", paymentInstruction.ContractNumber()),
			zap.Duration("duration", processingDuration))
		m.MetricsClient.Histogram(ctx, ports.MetricPaymentProcessingTimeInSeconds, processingDuration.Seconds(), successTags)
		m.MetricsClient.Count(ctx, successfulPaymentCounterName, 1, []string{paymentProviderTag})
	} else {
		zapctx.Debug(ctx, "flow_step #11b: Banking Circle payment failed for payment instruction %s with contract number %s in %v. Status was %v",
			zap.String("id", string(paymentInstruction.ID())),
			zap.String("merchant_contract_number", paymentInstruction.ContractNumber()),
			zap.Duration("duration", processingDuration),
			zap.String("status", string(bcProcessingStatus)))
		m.MetricsClient.Histogram(ctx, ports.MetricPaymentProcessingTimeInSeconds, processingDuration.Seconds(), failedTags)
	}
}

func (m observer) CouldntRequestBCRequest(ctx context.Context, id models.PaymentInstructionID, err error) {
	zapctx.Info(ctx, "[MakeBankingCirclePayment] (Execute) Could not create DTO to send to banking circle",
		zap.String("id", string(id)),
		zap.Error(err),
	)
}

func (m observer) MissingFunds(ctx context.Context, accountNumber, currency string) {
	zapctx.Info(ctx, "[CheckBankingCirclePaymentStatus] (Execute) Missing funds for account",
		zap.String("account_number", accountNumber))
	m.MetricsClient.Count(ctx, missingFunds, 1, []string{paymentProviderTag, currency})
}

func (m observer) CleanInput(ctx context.Context, oldFieldValue, newFieldValue, fieldName string) {
	zapctx.Info(ctx, "[createRequestDTO] Changing field value",
		zap.String("field_name", fieldName),
		zap.String("from", oldFieldValue),
		zap.String("to", newFieldValue))
	m.MetricsClient.Count(ctx, dataFieldCleaned, 1, []string{paymentProviderTag})
}
