package okta

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	jwtverifier "github.com/okta/okta-jwt-verifier-golang"
	oktaUtils "github.com/okta/samples-golang/custom-login/utils"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/adapters/http_server/middleware"
)

type Authentication struct {
	clientID     string
	clientSecret string
	issuer       string
	state        string
	redirectURI  string
}

const (
	nonceCookieName = "okta-nonce"
)

func NewAuthentication(clientID, clientSecret, redirectURI string) *Authentication {
	return &Authentication{
		clientID:     clientID,
		clientSecret: clientSecret,
		issuer:       "https://saltpayco.okta.com",
		state:        "foobar",
		redirectURI:  redirectURI,
	}
}

func (a Authentication) NewMiddleware() middleware.HTTPMiddleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenParts := strings.Split(authHeader, "Bearer ")
			bearerToken := tokenParts[1]

			toValidate := map[string]string{}
			toValidate["aud"] = "api://default" //???
			toValidate["cid"] = a.clientID

			jwtVerifierSetup := jwtverifier.JwtVerifier{
				Issuer:           a.issuer,
				ClaimsToValidate: toValidate,
			}

			_, err := jwtVerifierSetup.New().VerifyAccessToken(bearerToken)
			if err != nil {
				http.Error(w, "Troubles verifying access token", http.StatusBadRequest)
				return
			}

			handler.ServeHTTP(w, r)
		})
	}
}

func (a Authentication) AuthCodeCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("state") != a.state {
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("code") == "" {
		http.Error(w, "No auth code in callback handler (some Okta mishap)", http.StatusBadRequest)
		return
	}

	exchange := a.exchangeCode(r.URL.Query().Get("code"), r)

	// get nonce from cookie
	nonce, err := r.Cookie(nonceCookieName)
	if err != nil {
		http.Error(w, "Nonce not found", http.StatusBadRequest)
		return
	}

	_, verificationError := a.verifyToken(r.Context(), exchange.IDToken, nonce.Value)

	if verificationError != nil {
		http.Error(w, fmt.Sprintf("problem verifying token %+v", err), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func (a Authentication) LoginHandler(w http.ResponseWriter, r *http.Request) {
	nonce, _ := oktaUtils.GenerateNonce()
	var redirectPath string

	q := r.URL.Query()
	q.Add("client_id", a.clientID)
	q.Add("response_type", "code")
	q.Add("response_mode", "query")
	q.Add("scope", "openid profile email")
	q.Add("redirect_uri", a.redirectURI)
	q.Add("state", a.state)
	q.Add("nonce", nonce)

	redirectPath = a.issuer + "/v1/authorize?" + q.Encode()

	http.SetCookie(w, &http.Cookie{
		Name:    nonceCookieName,
		Value:   nonce,
		Expires: time.Now().UTC().Add(1 * time.Hour),
	})

	http.Redirect(w, r, redirectPath, http.StatusMovedPermanently)
}

func (a Authentication) exchangeCode(code string, r *http.Request) Exchange {
	authHeader := base64.StdEncoding.EncodeToString(
		[]byte(a.clientID + ":" + a.clientSecret))

	q := r.URL.Query()
	q.Add("grant_type", "authorization_code")
	q.Add("code", code)
	q.Add("redirect_uri", "http://localhost:8080/authorization-code/callback")

	url := a.issuer + "/v1/token?" + q.Encode()

	req, _ := http.NewRequest("POST", url, bytes.NewReader([]byte("")))
	h := req.Header
	h.Add("Authorization", "Basic "+authHeader)
	h.Add("Accept", "application/json")
	h.Add("Content-Type", "application/x-www-form-urlencoded")
	h.Add("Connection", "close")
	h.Add("Content-Length", "0")

	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var exchange Exchange
	err := json.Unmarshal(body, &exchange)
	if err != nil {
		return Exchange{}
	}

	return exchange
}

func (a Authentication) verifyToken(ctx context.Context, t string, nonce string) (*jwtverifier.Jwt, error) {
	tv := map[string]string{}
	tv["nonce"] = nonce
	tv["aud"] = a.clientID
	jv := jwtverifier.JwtVerifier{
		Issuer:           a.issuer,
		ClaimsToValidate: tv,
	}

	result, err := jv.New().VerifyIdToken(t)
	zapctx.Info(ctx, "token claims", zap.Any("claims", result.Claims))

	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	if result != nil {
		return result, nil
	}

	return nil, fmt.Errorf("token could not be verified: %s", "")
}
