package payment_store

import (
	"context"
	"fmt"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type LoggingAndMetricsPaymentObservability struct {
	metricsClient ports.MetricsClient
	callee        string
	logPrefix     string
}

var _ PaymentRepoObservability = (*LoggingAndMetricsPaymentObservability)(nil)

func NewLoggingAndMetricsPaymentObservabilityForPostgres(metricsClient ports.MetricsClient) *LoggingAndMetricsPaymentObservability {
	return &LoggingAndMetricsPaymentObservability{
		metricsClient: metricsClient,
		callee:        "postgres",
		logPrefix:     "[PaymentStore:postgres] ",
	}
}

const (
	piStoreMetricName          = "app_payment_instruction_store"
	piUpdateMetricName         = "app_payment_instruction_update"
	piStoreDurationMetricName  = "app_payment_instruction_store_duration"
	piUpdateDurationMetricName = "app_payment_instruction_update_duration"
	operationGet               = "get"
	operationStore             = "store"
	operationUpdate            = "update"
	operationGetReport         = "getreport"
	resultSuccess              = "success"
	resultFailure              = "failure"
)

func (l *LoggingAndMetricsPaymentObservability) GetSuccessful(ctx context.Context, _ models.PaymentInstructionID, responseTime int64) {
	tags := []string{l.callee, operationGet, resultSuccess}
	l.metricsClient.Histogram(ctx, piStoreDurationMetricName, float64(responseTime), tags)
	l.metricsClient.Count(ctx, piStoreMetricName, 1, tags)
}

func (l *LoggingAndMetricsPaymentObservability) FailedGet(ctx context.Context, id models.PaymentInstructionID, err error) {
	tags := []string{l.callee, operationGet, resultFailure}
	l.metricsClient.Count(ctx, piStoreMetricName, 1, tags)

	zapctx.Error(ctx, fmt.Sprintf("%sfailed to get payment instruction", l.logPrefix),
		zap.String("id", string(id)),
		zap.Error(err))
}

func (l *LoggingAndMetricsPaymentObservability) StoreSuccessful(ctx context.Context, _ models.PaymentInstructionID, responseTime int64) {
	tags := []string{l.callee, operationStore, resultSuccess}
	l.metricsClient.Histogram(ctx, piStoreDurationMetricName, float64(responseTime), tags)
	l.metricsClient.Count(ctx, piStoreMetricName, 1, tags)
}

func (l *LoggingAndMetricsPaymentObservability) FailedStore(ctx context.Context, id models.PaymentInstructionID, contractNumber string, err error) {
	tags := []string{l.callee, operationStore, resultFailure}
	l.metricsClient.Count(ctx, piStoreMetricName, 1, tags)
	zapctx.Error(ctx, fmt.Sprintf("%sfailed to store payment instruction", l.logPrefix),
		zap.String("id", string(id)),
		zap.String("merchant_contract_number", contractNumber),
		zap.Error(err))
}

func (l *LoggingAndMetricsPaymentObservability) FailedUpdate(ctx context.Context, id models.PaymentInstructionID, err error) {
	tags := []string{l.callee, operationUpdate, resultFailure}
	l.metricsClient.Count(ctx, piUpdateMetricName, 1, tags)
	zapctx.Error(ctx, fmt.Sprintf("%sfailed to update payment instruction", l.logPrefix),
		zap.String("id", string(id)),
		zap.Error(err))
}

func (l *LoggingAndMetricsPaymentObservability) UpdateSuccessful(ctx context.Context, _ models.PaymentInstructionID, responseTime int64) {
	tags := []string{l.callee, operationUpdate, resultSuccess}
	l.metricsClient.Histogram(ctx, piUpdateDurationMetricName, float64(responseTime), tags)
	l.metricsClient.Count(ctx, piUpdateMetricName, 1, tags)
}

func (l *LoggingAndMetricsPaymentObservability) ReceivedStoreInstruction(context.Context, models.PaymentInstructionID) {
	// todo: metric here?
}

func (l *LoggingAndMetricsPaymentObservability) ReceivedGetInstruction(context.Context, models.PaymentInstructionID) {
	// todo: metric here?
}

func (l *LoggingAndMetricsPaymentObservability) PaymentInstructionNotFound(ctx context.Context, id models.PaymentInstructionID) {
	zapctx.Debug(ctx, fmt.Sprintf("%spayment instruction not found", l.logPrefix),
		zap.String("id", string(id)))
}

func (l *LoggingAndMetricsPaymentObservability) GotReport(ctx context.Context, duration int64) {
	tags := []string{l.callee, operationGetReport, resultSuccess}
	l.metricsClient.Histogram(ctx, piStoreDurationMetricName, float64(duration), tags)
	l.metricsClient.Count(ctx, piStoreMetricName, 1, tags)
}
