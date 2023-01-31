package adminsdk

import (
	"context"
	"fmt"
	"net/http"
)

const (
	toggleOn  = "%sadmin/features/%s/toggle/on"
	toggleOff = "%sadmin/features/%s/toggle/off"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type AdminSDK struct {
	unleashServiceURL string
	serviceToken      string
	httpClient        HTTPClient
}

func NewAdminSDK(
	unleashServiceURL string,
	serviceToken string,
	client HTTPClient,
) *AdminSDK {
	return &AdminSDK{
		unleashServiceURL: unleashServiceURL,
		serviceToken:      serviceToken,
		httpClient:        client,
	}
}

// FeatureOn toggles on a certain feature. Assumes feature already exists within unleash.
func (a *AdminSDK) FeatureOn(ctx context.Context, feature string) error {
	path := fmt.Sprintf(toggleOn, a.unleashServiceURL, feature)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("error creating feature (%s) on request: %w", feature, err)
	}
	req.Header.Set("Authorization", a.serviceToken)

	res, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request for feature (%s) on: %w", feature, err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("error status enabling feature (%s). status %d", feature, res.StatusCode)
	}
	return nil
}

// FeatureOff toggles on a certain feature. Assumes feature already exists within unleash.
func (a *AdminSDK) FeatureOff(ctx context.Context, feature string) error {
	path := fmt.Sprintf(toggleOff, a.unleashServiceURL, feature)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("error creating feature (%s) off request: %w", feature, err)
	}
	req.Header.Set("Authorization", a.serviceToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request for feature (%s) off: %w", feature, err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("error status disabling feature (%s). status %d", feature, res.StatusCode)
	}
	return nil
}
