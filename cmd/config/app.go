package config

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/adapters/env"
	"github.com/saltpay/settlements-payments-system/internal/projectpath"
)

const (
	UnprocessedQueueVisibilityTimeoutSeconds = 10 * 60
	UnprocessedQueueReceiveBatchSize         = 1
	UncheckedQueueVisibilityTimeoutSeconds   = 25 * 60
	UncheckedQueueReceiveBatchSize           = 1
)

// Config represents the config used by the app.
type Config struct {
	EnvName       env.EnvName `split_words:"true"`
	ComponentName string      `split_words:"true"`
	LogLevel      string      `split_words:"true"`
	AwsRegion     string      `split_words:"true"`
	AwsEndpoint   string      `split_words:"true"`

	// DD_AGENT_HOST is provided to the application pod by the Kubernetes deployment
	DdAgentHost string `split_words:"true"`

	DdAgentPort                               int           `split_words:"true"`
	Port                                      string        `split_words:"true"`
	AwsDisableSSL                             bool          `split_words:"true"`
	S3UfxPaymentFilesBucketName               string        `split_words:"true"`
	SqsDefaultVisibilityTimeoutSeconds        int64         `split_words:"true"`
	SqsDefaultReceiveWaitTimeSeconds          int64         `split_words:"true"`
	SqsDefaultReceiveBatchSize                int64         `split_words:"true"`
	SqsUfxFileNotificationQueueName           string        `split_words:"true"`
	SqsUfxFileNotificationDLQName             string        `split_words:"true"`
	SqsBankingCircleUnprocessedQueueName      string        `split_words:"true"`
	SqsBankingCircleUnprocessedDLQName        string        `split_words:"true"`
	SqsBankingCircleProcessedQueueName        string        `split_words:"true"`
	SqsBankingCircleProcessedDLQName          string        `split_words:"true"`
	SqsBankingCircleUncheckedQueueName        string        `split_words:"true"`
	SqsBankingCircleUncheckedDLQName          string        `split_words:"true"`
	SqsIslandsbankiUnprocessedQueueName       string        `split_words:"true"`
	SqsIslandsbankiUnprocessedDLQName         string        `split_words:"true"`
	SqsNumberOfDeleteRetries                  int           `split_words:"true"`
	SqsSleepTime                              time.Duration `split_words:"true"`
	SqsAllowPurge                             bool          `split_words:"true"`
	UfxProcessingMaxMessages                  int64         `split_words:"true"`
	BankingCircleAPIAuthorizationBaseURL      string        `split_words:"true"`
	BankingCircleAPIBaseURL                   string        `split_words:"true"`
	BankingCircleAPISecretName                string        `split_words:"true"`
	BankingCircleAPITimeout                   time.Duration `split_words:"true"`
	BankingCircleAPIInsecureSkipVerify        bool          `split_words:"true"`
	BankingCircleTokenIntervalBeforeExpire    time.Duration `split_words:"true"`
	BankingCircleStatusCheckDelay             int64         `split_words:"true"`
	BankingCircleMakePaymentDelayMilliseconds int64         `split_words:"true"`
	BankingCircleMaxCheckIterations           int64         `split_words:"true"`
	BankingCircleMakePaymentWorkerPoolSize    int64         `split_words:"true"`
	BankingCircleCheckPaymentWorkerPoolSize   int64         `split_words:"true"`
	FeatureFlagServiceURL                     string        `split_words:"true"`
	FeatureFlagServiceAPIKeySecretName        string        `split_words:"true"`
	UseFakeBankingCircleAPI                   bool          `split_words:"true"`
	OktaCredentialsSecretKeyName              string        `split_words:"true"`
	OktaRedirectURI                           string        `split_words:"true"`
	BearerAuthSecretKeyName                   string        `split_words:"true"`
	PermittedTestUsers                        []string      `split_words:"true"`
	PostgresConnectionSecretName              string        `split_words:"true"`
	NetworkingCheckAddress                    []string      `split_words:"true"`
	HealthCheckTimeout                        time.Duration `split_words:"true"`
	HealthCheckInterval                       time.Duration `split_words:"true"`
	FeatureFlagServiceAdminAPIKeySecretName   string        `split_words:"true"`
	FailedPaymentsThreshold                   int           `split_words:"true"`
	EnablePostPaymentEndpoint                 bool          `split_words:"true"`
	Kafka                                     KafkaConfig
}

type KafkaConfig struct {
	Endpoint           []string `split_words:"true"`
	UsernameSecretName string   `split_words:"true"`
	PasswordSecretName string   `split_words:"true"`
	Topics             KafkaTopics
}

type KafkaTopics struct {
	Transactions                    string `split_words:"true"`
	AcquiringHostTransactionUpdates string `split_words:"true"`
	UnprocessedISBPayments          string `split_words:"true"`
	PaymentStateUpdates             string `split_words:"true"`
}

// LoadConfig loads the app config from environment variables.
func LoadConfig(ctx context.Context) (Config, error) {
	LoadEnvVarsFromFile(ctx) // this is a temporary measure till we move to SaltPay infrastructure

	// load Config from env vars
	var config Config
	err := envconfig.Process("", &config)
	if err != nil {
		return Config{}, err
	}

	zapctx.Info(ctx, "app config loaded", zap.Any("config", config))
	return config, nil
}

func LoadEnvVarsFromFile(ctx context.Context) {
	// look up the env name from env variables; integration/production/local
	envName := os.Getenv("ENV_NAME")
	if envName == "" {
		panic("could not find ENV_NAME environment variable")
	}

	// load env vars from `.envName.env`
	fileName := fmt.Sprintf(".%s.env", envName)

	filePath := path.Join(projectpath.Root, fileName)

	fmt.Println("===filePath", filePath)
	err := godotenv.Load(filePath)
	if err != nil {
		zapctx.Fatal(ctx, "failed to load env vars from file", zap.Error(err))
	}
}

func TKI() string {
	return "settlements-payments-system"
}
