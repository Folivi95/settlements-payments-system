package unleash

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	unleashlib "github.com/Unleash/unleash-client-go/v3"

	"github.com/saltpay/settlements-payments-system/internal/adapters/env"
	"github.com/saltpay/settlements-payments-system/internal/adapters/unleash/adminsdk"

	zapctx "github.com/saltpay/go-zap-ctx"
)

type UnleashFeatureFlagService struct {
	client   *unleashlib.Client
	adminSdk *adminsdk.AdminSDK
	envName  string
}

func NewUnleashFeatureFlagService(ctx context.Context, apiURL string, apiKey string, envName string, adminSdk *adminsdk.AdminSDK) (UnleashFeatureFlagService, error) {
	client, err := unleashlib.NewClient(
		unleashlib.WithUrl(apiURL),
		unleashlib.WithAppName("settlements-payments-system"),
		unleashlib.WithEnvironment(envName),
		unleashlib.WithCustomHeaders(http.Header{"Authorization": {apiKey}}),
	)
	if err != nil {
		return UnleashFeatureFlagService{}, err
	}

	// https://github.com/Unleash/unleash-client-go/tree/ebb5075757466f44d43973a372754ece6ed7d490#caveat
	// This client uses go routines to report several events and doesn't drain the channel by default.
	// So you need to either register a listener using WithListener or drain the channel "manually" (demonstrated in this example).
	zapctx.Info(ctx, "about to start Feature Flag channel draining loop in go routine")
	go func(ctx context.Context) {
		for {
			select {
			case e := <-client.Errors():
				zapctx.Info(ctx, "unleash error", zap.Error(e))
			case w := <-client.Warnings():
				zapctx.Info(ctx, "unleash warning", zap.Error(w))
			case <-client.Ready():
				zapctx.Info(ctx, "unleash ready")
			case <-client.Count():
				continue
			case <-client.Sent():
				continue
			case cd := <-client.Registered():
				zapctx.Info(ctx, "unleash registered", zap.Any("client_data", cd))
			}
		}
	}(ctx)

	if envName == string(env.Tilt) {
		envName = string(env.Local)
	}

	return UnleashFeatureFlagService{
		client:   client,
		adminSdk: adminSdk,
		envName:  envName,
	}, nil
}

func (uffs UnleashFeatureFlagService) IsIngestionEnabledFromBankingCircleUnprocessedQueue() bool {
	return uffs.client.IsEnabled(uffs.envName + "---settlements_payments---is_ingestion_enabled_from_banking_circle_unprocessed_queue")
}

func (uffs UnleashFeatureFlagService) IsIngestionEnabledFromBankingCircleUncheckedQueue() bool {
	return uffs.client.IsEnabled(uffs.envName + "---settlements_payments---is_ingestion_enabled_from_banking_circle_unchecked_queue")
}

func (uffs UnleashFeatureFlagService) IsKafkaIngestionEnabledForPaymentTransactions() bool {
	return uffs.client.IsEnabled(uffs.envName + "---settlements_payments---is_kafka_ingestion_enabled_for_payment_transactions")
}

func (uffs UnleashFeatureFlagService) IsKafkaPublishingEnableForAcquiringHostTransactions() bool {
	return uffs.client.IsEnabled(uffs.envName + "---settlements_payments---is_kafka_ingestion_enabled_for_payment_transactions_updates")
}

func (uffs UnleashFeatureFlagService) IsIngestionEnabledFromUfxFileNotificationQueue() bool {
	return uffs.client.IsEnabled(uffs.envName + "---settlements_payments---is_ingestion_enabled_from_ufx_file_notification_queue")
}

func (uffs UnleashFeatureFlagService) ToggleOffIngestionFromBankingCirclePayments(ctx context.Context) error {
	return uffs.adminSdk.FeatureOff(ctx, uffs.envName+"---settlements_payments---is_ingestion_enabled_from_banking_circle_unprocessed_queue")
}
