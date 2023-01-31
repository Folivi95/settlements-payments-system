package testdoubles

import "context"

type FeatureFlagService struct{}

func (d FeatureFlagService) IsIngestionEnabledFromBankingCircleUnprocessedQueue() bool {
	return true
}

func (d FeatureFlagService) IsIngestionEnabledFromBankingCircleUncheckedQueue() bool {
	return true
}

func (d FeatureFlagService) IsKafkaIngestionEnabledForPaymentTransactions() bool {
	return true
}

func (d FeatureFlagService) IsKafkaPublishingEnableForAcquiringHostTransactions() bool {
	return true
}

func (d FeatureFlagService) IsIngestionEnabledFromUfxFileNotificationQueue() bool {
	return true
}

func (d FeatureFlagService) ToggleOffIngestionFromBankingCirclePayments(context.Context) error {
	return nil
}
