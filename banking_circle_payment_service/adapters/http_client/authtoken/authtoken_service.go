package authtoken

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type AuthTokenService struct {
	currentAuthToken string
	authConfig       AuthConfig
	httpClient       *http.Client
	metricsClient    ports.MetricsClient
	timer            *time.Timer
	authURL          string
}

type AuthConfig struct {
	AuthorizationBaseURL      string
	APIUsername               string
	APIPassword               string
	TokenIntervalBeforeExpire time.Duration
	Scheduler                 Scheduler
}

type AuthResponseDto struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

func NewAuthTokenService(
	authConfig AuthConfig,
	client *http.Client,
	metrics ports.MetricsClient,
) *AuthTokenService {
	if authConfig.Scheduler == nil {
		authConfig.Scheduler = SchedulerFunc(time.AfterFunc)
	}

	authToken := AuthTokenService{
		currentAuthToken: "",
		authConfig:       authConfig,
		httpClient:       client,
		metricsClient:    metrics,
		timer:            time.NewTimer(time.Second),
		authURL:          authConfig.AuthorizationBaseURL + "/authorizations/authorize",
	}

	return &authToken
}

const (
	clientResponseMetricName = "app_http_client_resp_time_ms"
)

func (b *AuthTokenService) GetAccessToken(ctx context.Context) (string, error) {
	if b.currentAuthToken == "" {
		err := b.requestToken(ctx)
		if err != nil {
			return "", errors.Wrap(err, "failed to authorize in the background")
		}
	}

	return b.currentAuthToken, nil
}

func (b *AuthTokenService) requestToken(ctx context.Context) error {
	err := b.authorize(ctx)
	if err != nil {
		b.currentAuthToken = ""
		zapctx.Warn(ctx, "failed to request token", zap.Error(err))
		return err
	}

	return nil
}

func (b *AuthTokenService) authorize(ctx context.Context) error {
	authConfig := b.authConfig
	req, err := http.NewRequest(http.MethodGet, b.authURL, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create a request")
	}
	req.SetBasicAuth(authConfig.APIUsername, authConfig.APIPassword)

	startTime := time.Now()

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "failed to request a new token")
	}
	defer resp.Body.Close()

	responseTime := time.Since(startTime).Milliseconds()
	statusCode := fmt.Sprintf("%dxx", resp.StatusCode/100)
	tags := []string{"banking_circle", "get_auth_token", statusCode}
	b.metricsClient.Histogram(ctx, clientResponseMetricName, float64(responseTime), tags)

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response body")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authorize call did not return a 200 OK; error code: %d, error body: %s", resp.StatusCode, string(respBytes))
	}

	var authResp AuthResponseDto
	err = json.Unmarshal(respBytes, &authResp)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal the response")
	}

	b.currentAuthToken = authResp.AccessToken

	tokenExpirationSeconds, err := strconv.ParseInt(authResp.ExpiresIn, 10, 64)
	if err != nil {
		return errors.Wrap(err, "failed to convert expiration time")
	}

	nextAuthorizationAfterDuration := time.Second * time.Duration(tokenExpirationSeconds)
	nextAuthorizationAfterDuration -= b.authConfig.TokenIntervalBeforeExpire
	if nextAuthorizationAfterDuration < 0 {
		return errors.Wrap(err, "next authorization schedule can not be negative duration")
	}

	b.authConfig.Scheduler.AfterFunc(nextAuthorizationAfterDuration, func() {
		_ = b.requestToken(context.Background())
	})

	return nil
}
