//go:generate moq -out mocks/feature_flag_service_moq.go -pkg=mocks . FeatureFlagService

package ports

import "context"

type FeatureFlagService interface {
	IsIngestionEnabledFromBankingCircleUnprocessedQueue() bool
	IsIngestionEnabledFromBankingCircleUncheckedQueue() bool
	IsKafkaIngestionEnabledForPaymentTransactions() bool
	IsKafkaPublishingEnableForAcquiringHostTransactions() bool
	IsIngestionEnabledFromUfxFileNotificationQueue() bool
	ToggleOffIngestionFromBankingCirclePayments(ctx context.Context) error
}
