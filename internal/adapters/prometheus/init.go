package prometheus

import (
	"context"
	"strings"

	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

const (
	// bcProvider defines the provider that we currently have
	// TODO: This should be improved to get a list of all of our providers when we start to support more providers.
	bcProvider  = "banking_circle"
	isbProvider = "islandsbanki"
	// resultSuccess defines payment success event.
	resultSuccess = "success"
	// resultFailure defines payment failure event.
	resultFailure = "failure"
	// source defines where the payment instruction comes from.
	way4Source  = "way4_ufx"
	solarSource = "solar_event"
)

// initMetrics function should set the initial values for metrics were we want to alert when it changes from 0 to some other
// value.
// This allows us to prevent issues where a metrics doesn't exist prior to first increment.
func (m *MetricsClient) initMetrics(ctx context.Context) {
	// init missing funds per currency metric.
	m.initMissingFundsMetricZeroValue(ctx)
	// init missing account funds per currency metric.
	m.initMissingAccountFundsMetricZeroValue(ctx)
	// init missing account funds using loan per currency metric.
	m.initMissingAccountFundsUsingLoanMetricZeroValue(ctx)
	// init request payment metric.
	m.initRequestPaymentMetricZeroValue(ctx)
	// init check payment status metric.
	m.initCheckPaymentStatusMetricZeroValue(ctx)
	// init payments count per currency metric.
	m.initPaymentsCountPerCurrencyMetricZeroValue(ctx)
	// init payments amount per currency metric.
	m.initPaymentsAmountPerCurrencyMetricZeroValue(ctx)
	// init payment instructions received metric.
	m.initPaymentInstructionsReceivedMetricZeroValue(ctx)
	// init payment successful metric.
	m.initAppSettlementsProviderSuccessPayment(ctx)
	// init incorrect account number format metric.
	m.initAppSettlementsIncorrectAccountNumberFormatZeroValue(ctx)
	// init payment state update metric at zero.
	m.initPaymentStateUpdateMetricZeroValue(ctx)
}

func (m *MetricsClient) initMissingFundsMetricZeroValue(ctx context.Context) {
	m.initMetric(ctx, "app_settlements_provider_missing_funds", true, bcProvider)
	m.initMetric(ctx, "app_settlements_provider_missing_funds", true, isbProvider)
}

func (m *MetricsClient) initMissingAccountFundsMetricZeroValue(ctx context.Context) {
	m.initMetric(ctx, "app_settlements_provider_missing_account_funds", true)
}

func (m *MetricsClient) initMissingAccountFundsUsingLoanMetricZeroValue(ctx context.Context) {
	m.initMetric(ctx, "app_settlements_provider_missing_account_funds_using_loan", true)
}

func (m *MetricsClient) initRequestPaymentMetricZeroValue(ctx context.Context) {
	m.initMetric(ctx, "app_settlements_provider_request_payment", true, resultSuccess, bcProvider)
	m.initMetric(ctx, "app_settlements_provider_request_payment", true, resultFailure, bcProvider)
	m.initMetric(ctx, "app_settlements_provider_request_payment", true, resultSuccess, isbProvider)
	m.initMetric(ctx, "app_settlements_provider_request_payment", true, resultFailure, isbProvider)
}

func (m *MetricsClient) initCheckPaymentStatusMetricZeroValue(ctx context.Context) {
	m.initMetric(ctx, "app_settlements_provider_check_payment_status", true, resultSuccess, bcProvider)
	m.initMetric(ctx, "app_settlements_provider_check_payment_status", true, resultFailure, bcProvider)
	m.initMetric(ctx, "app_settlements_provider_check_payment_status", true, resultSuccess, isbProvider)
	m.initMetric(ctx, "app_settlements_provider_check_payment_status", true, resultFailure, isbProvider)
}

func (m *MetricsClient) initPaymentsCountPerCurrencyMetricZeroValue(ctx context.Context) {
	m.initMetric(ctx, "app_settlements_payments_count", true, bcProvider)
	m.initMetric(ctx, "app_settlements_payments_count", true, isbProvider)
}

func (m *MetricsClient) initPaymentsAmountPerCurrencyMetricZeroValue(ctx context.Context) {
	m.initMetric(ctx, "app_settlements_payments_amount", true, bcProvider)
	m.initMetric(ctx, "app_settlements_payments_amount", true, isbProvider)
}

func (m *MetricsClient) initPaymentInstructionsReceivedMetricZeroValue(ctx context.Context) {
	m.initMetric(ctx, "app_payment_instructions_received", false, way4Source)
	m.initMetric(ctx, "app_payment_instructions_received", false, solarSource)
}

func (m *MetricsClient) initAppSettlementsProviderSuccessPayment(ctx context.Context) {
	m.initMetric(ctx, "app_settlements_provider_success_payment", false, bcProvider)
	m.initMetric(ctx, "app_settlements_provider_success_payment", false, isbProvider)
}

func (m *MetricsClient) initAppSettlementsIncorrectAccountNumberFormatZeroValue(ctx context.Context) {
	m.initMetric(ctx, "app_settlements_payments_incorrect_account_number", true)
}

func (m *MetricsClient) initPaymentStateUpdateMetricZeroValue(ctx context.Context) {
	m.initMetric(ctx, "app_payment_instruction_update", true)
}

// initMetric will get the metric for the metricName with the labelsValues (applied in order) before the currency label value.
// todo: refactor this function as some metrics don't need the currency tag
func (m *MetricsClient) initMetric(ctx context.Context, metricName string, needsCurrencyTag bool, labelValues ...string) {
	if needsCurrencyTag {
		currencies := models.GetAllCurrencyCodes()
		for _, c := range currencies {
			counter := m.counters[metricName]
			labels := append(labelValues, string(c))
			metric, err := counter.GetMetricWithLabelValues(labels...)
			if err != nil {
				zapctx.Error(ctx, "unable to init metric",
					zap.String("metric_name", metricName),
					zap.String("currency", string(c)),
					zap.String("tags", strings.Join(labelValues, ", ")),
					zap.Error(err),
				)

				continue
			}
			// set the default value to zero
			metric.Add(0)
		}
	} else {
		counter := m.counters[metricName]
		metric, err := counter.GetMetricWithLabelValues(labelValues...)
		if err != nil {
			zapctx.Error(ctx, "unable to init metric",
				zap.String("metric_name", metricName),
				zap.String("tags", strings.Join(labelValues, ", ")),
				zap.Error(err),
			)
		}
		// set the default value to zero
		metric.Add(0)
	}
}
