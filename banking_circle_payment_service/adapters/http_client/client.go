package http_client

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/adapters/http_client/authtoken"

	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

func NewBankingCircleHTTPClient(apiConfig BankingCircleAPIConfig, metricsClient ports.MetricsClient) (*http.Client, error) {
	cert, err := tls.X509KeyPair([]byte(apiConfig.ClientCertificatePublicKey), []byte(apiConfig.ClientCertificatePrivateKey))
	if err != nil {
		return nil, err
	}

	secureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates:       []tls.Certificate{cert},
			Renegotiation:      tls.RenegotiateFreelyAsClient,
			InsecureSkipVerify: apiConfig.InsecureSkipVerify,
		},
	}

	authConfig := authtoken.AuthConfig{
		AuthorizationBaseURL:      apiConfig.AuthorizationBaseURL,
		APIUsername:               apiConfig.APIUsername,
		APIPassword:               apiConfig.APIPassword,
		TokenIntervalBeforeExpire: apiConfig.TokenIntervalBeforeExpire,
	}

	httpTimeout := apiConfig.Timeout * time.Second

	tokenHTTPClient := &http.Client{
		Transport: secureTransport,
		Timeout:   httpTimeout,
	}

	authTokenService := authtoken.NewAuthTokenService(authConfig, tokenHTTPClient, metricsClient)

	bcHTTPClient := &http.Client{
		Transport: &bcTransport{
			delegate:     secureTransport,
			getAuthToken: authTokenService.GetAccessToken,
		},
		Timeout: httpTimeout,
	}
	return bcHTTPClient, nil
}

type BankingCircleAPIConfig struct {
	AuthorizationBaseURL        string
	BaseURL                     string
	APIUsername                 string
	APIPassword                 string
	ClientCertificatePublicKey  string
	ClientCertificatePrivateKey string
	Timeout                     time.Duration
	TokenIntervalBeforeExpire   time.Duration
	InsecureSkipVerify          bool
}

type bcTransport struct {
	delegate     http.RoundTripper
	getAuthToken func(ctx context.Context) (string, error)
}

func (t *bcTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Content-Type", "application/json")
	authToken, err := t.getAuthToken(req.Context())
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+authToken)

	return t.delegate.RoundTrip(req)
}
