//go:build integration
// +build integration

package integration_testing

import (
	"os"
	"testing"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/adapters/http_client"
	"github.com/saltpay/settlements-payments-system/internal/adapters/testdoubles"
)

func TestBankingCircleAPIClientAgainstFakes(t *testing.T) {
	t.Run("CheckAccountBalance", func(t *testing.T) {
		t.Run("returns account balance for EUR", func(t *testing.T) {
			var (
				accountNumber = uuid.New().String()
				apiClient     = newBankingCircleFakeClient(t)
			)

			accountBalance, err := apiClient.CheckAccountBalance(accountNumber)
			assert.NoError(t, err)

			assert.Equal(t, 1000000000.00, accountBalance.Result[0].IntraDayAmount)
			assert.Equal(t, 1000000000.00, accountBalance.Result[0].BeginOfDayAmount)
			assert.Equal(t, models.CurrencyCode("EUR"), accountBalance.Result[0].Currency)
		})
	})
}

func newBankingCircleFakeClient(t *testing.T) *http_client.BankingCircleAPIClient {
	var (
		fakeBCBaseURL = os.Getenv("FAKE_BC_BASE_URL")
		dummyMetrics  = testdoubles.DummyMetricsClient{}
	)

	config := http_client.BankingCircleAPIConfig{
		AuthorizationBaseURL:        fakeBCBaseURL,
		BaseURL:                     fakeBCBaseURL,
		APIUsername:                 uuid.New().String(),
		APIPassword:                 uuid.New().String(),
		ClientCertificatePublicKey:  getFakeCertPubliKey(),
		ClientCertificatePrivateKey: getFakeCertPrivateKey(),
		Timeout:                     30,
		InsecureSkipVerify:          true,
	}

	httpClient, err := http_client.NewBankingCircleHTTPClient(config, dummyMetrics)
	assert.NoError(t, err)

	client, err := http_client.NewAPIClient(httpClient, fakeBCBaseURL, dummyMetrics)
	assert.NoError(t, err)

	return client
}

func getFakeCertPrivateKey() string {
	return `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDNCdS+Kx4KOww6
GLpdgi/5Hjgglb218VOwcIII4BStsshCe85Mgz77pdnJSgKQKVbbF/yZC5QM9get
h5F9qEqYaVCKNUaFYBGGD2v7LTU7e7VD82iN8JvlsvaYPbj5Q+uSaTkNoenWaAXO
NKrFoNXXVGKofWK3j2NJEmcPGP+4MeyRMAT5/cgN9Q6YnZTZzlxtoD1ofmHXV95g
7KWQiTeoMOtGVZUeOlK/XR0v3J8uycfx8SxtchV2DLgFgtwTmrQnW4mAfsQ/Wh74
THZN9K7M9C5xHfNoFPxQ/OE82XsqS0PBGM/lB83jeQuwmH6ATx6UkMyqkdZ1OPdp
kU6CBZ1fAgMBAAECggEADxYIFy3o+eu6TJQBMlwf136Htq4N1VM6SyMcDjcejmE3
Jt0hIrQNcEqVqZ/ObHj+MQSky0X00LdRfU0aQVqeknQ9Ps4IsEuPPoPn+AUtg4Do
p2VDbh4j+lSenDj+YSjELnObhQtCv0nME44AeqYI1d0ZnTgMiWD9dyTpfEzkk9LG
x6Mp3kdo+T4PbJdEnoHTvmCg9CQBPqOOvz8heKg4tU2Ovbv/e//Zxqv3Pdys1D4F
6v2Pw4Vto3/dmHX096ZVCjI9+XQwEkK9FkIdSCq7tIiBSvFs0jt/wX4atN0q+WPa
92rTTAfpKjbu32Lo8zrPh+GOd48iISQMl4+kIm+uAQKBgQDx3ThVFeskq6g8YLRB
EQEUgAX0WPgzC+oplVXgnBkrlD4MrnSynlaZXUERkzIU1L96sTNoM6cwZ5+w+cE1
ArTNZrXacOxtbCYeeiF8Gg+T1wM3JtZQ2NrclqwLWWfjYezpdigIC4xTp6IKC4aY
Dym3webjkkpsgU4lnz7mo0vS3wKBgQDZBZ96TLLMH1oO6zZ0tk2IeYc1YGHm4Gja
w9gjXDEduqMgQAdB2W9BsM16DF843Oo6FjDTWh06lpMxtoSF9W9C/rHHf5U7McQ6
oGZ2axuVG04zhgVkJNFTLCXtnxIKG+NgmJG2LZ50EZM0s09BHL7lxwhJPBU1D4O7
npA3+o4FgQKBgGZ3d2csuws1Ijg6LAOo5ZE+z8b+bmCJ+rGVT+WxnERHMKaEvnHx
/PRKesesWbpTi6+6JPJPd9RdAl2i4gTIWbrvebnKv494Ewo0ab0++TyEChuye3eS
994eg1LnlMjTcuBRq5IE+nVyfobM7T+8pMrx/hSJpLgla+sqdSaXJgd3AoGBAJXs
74aS0/Z2NkYWMpGgm6GLq1+xjRDtuSJgp8GN4BSUqjsOYLUaHGU6WklVoLbszxd/
2w03tPeTrG5sk9LjgpC62WBkAFlbgR+rTf3C8tQof/bSQIk1cjLOTgmBmfnH2GYU
IJ3FmDDBL3v53+ewjyS4Qj4ttszoQe5slV9GxbSBAoGAZA1aNkneHex8Y27hQIYT
NXxkn/IDZawzpt/TFSSyvjWA5EMZo3bZiuKNwR6vgtiql4N1IzlpVLEPxPxgVRjO
cT3UNzoIkrVqObi2PJ+4frCxEQQcESzFbuJXVRRiJjieV6+LqUbrMx1qYLP8FG0y
vH1KFpM/jH9JgF7xFQoyu74=
-----END PRIVATE KEY-----`
}

func getFakeCertPubliKey() string {
	return `-----BEGIN CERTIFICATE-----
MIIFFDCCA/ygAwIBAgIMHJgbykjL61W4UGeHMA0GCSqGSIb3DQEBCwUAMFsxCzAJ
BgNVBAYTAkJFMRkwFwYDVQQKExBHbG9iYWxTaWduIG52LXNhMTEwLwYDVQQDEyhH
bG9iYWxTaWduIEdDQyBSMyBQZXJzb25hbFNpZ24gMSBDQSAyMDIwMB4XDTIxMDUx
NDEzMDI0NloXDTI0MDUxNDEzMDI0NlowTjEiMCAGA1UEAwwZYWthc2gua3VyZGVr
YXJAc2FsdHBheS5jbzEoMCYGCSqGSIb3DQEJARYZYWthc2gua3VyZGVrYXJAc2Fs
dHBheS5jbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAM0J1L4rHgo7
DDoYul2CL/keOCCVvbXxU7BwggjgFK2yyEJ7zkyDPvul2clKApApVtsX/JkLlAz2
B62HkX2oSphpUIo1RoVgEYYPa/stNTt7tUPzaI3wm+Wy9pg9uPlD65JpOQ2h6dZo
Bc40qsWg1ddUYqh9YrePY0kSZw8Y/7gx7JEwBPn9yA31DpidlNnOXG2gPWh+YddX
3mDspZCJN6gw60ZVlR46Ur9dHS/cny7Jx/HxLG1yFXYMuAWC3BOatCdbiYB+xD9a
HvhMdk30rsz0LnEd82gU/FD84TzZeypLQ8EYz+UHzeN5C7CYfoBPHpSQzKqR1nU4
92mRToIFnV8CAwEAAaOCAeMwggHfMA4GA1UdDwEB/wQEAwIFoDCBowYIKwYBBQUH
AQEEgZYwgZMwTgYIKwYBBQUHMAKGQmh0dHA6Ly9zZWN1cmUuZ2xvYmFsc2lnbi5j
b20vY2FjZXJ0L2dzZ2NjcjNwZXJzb25hbHNpZ24xY2EyMDIwLmNydDBBBggrBgEF
BQcwAYY1aHR0cDovL29jc3AuZ2xvYmFsc2lnbi5jb20vZ3NnY2NyM3BlcnNvbmFs
c2lnbjFjYTIwMjAwTAYDVR0gBEUwQzBBBgkrBgEEAaAyASgwNDAyBggrBgEFBQcC
ARYmaHR0cHM6Ly93d3cuZ2xvYmFsc2lnbi5jb20vcmVwb3NpdG9yeS8wCQYDVR0T
BAIwADBJBgNVHR8EQjBAMD6gPKA6hjhodHRwOi8vY3JsLmdsb2JhbHNpZ24uY29t
L2dzZ2NjcjNwZXJzb25hbHNpZ24xY2EyMDIwLmNybDAkBgNVHREEHTAbgRlha2Fz
aC5rdXJkZWthckBzYWx0cGF5LmNvMB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEF
BQcDBDAfBgNVHSMEGDAWgBSFu/DMxDa1CmJ2o5kuj7s6aq3FUTAdBgNVHQ4EFgQU
4AkSBCO8OQcvd6Vp0baayd+spxQwDQYJKoZIhvcNAQELBQADggEBAAWiEMjRmPZy
EFwlh7DqtTw6zvSBYtT/bqeAH0sQVv/3LoRCaeibwyBsaZ+z1XGOPKLLfEZOdU2S
xmu5jtfSfLCcvErmAXKJ9p09GfU//rLz+K2aJnFBPVRjQMOmgSUizqnfFNoEVtIP
kSo/+k9XuAjgKwe7yfms30x1+AtuaPNEiCmimJm6SpampYQruS1jRCMZ9GAhoqev
1r35CJmAlbskS0gem99f5aoi9Mw3hrmvzPvIGrp0hpHHpDjwT4uIREPCZ+zMbx+z
ytKdphm9/Tid9sMUyHiLtCzUneKC6Dzuaw/nFxV3iVR29O10o2TkYAMx68fpPF7N
fsxZ9x6t/9E=
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIEvDCCA6SgAwIBAgIQeEqpEhjRpCYIUTzTZlVDozANBgkqhkiG9w0BAQsFADBM
MSAwHgYDVQQLExdHbG9iYWxTaWduIFJvb3QgQ0EgLSBSMzETMBEGA1UEChMKR2xv
YmFsU2lnbjETMBEGA1UEAxMKR2xvYmFsU2lnbjAeFw0yMDA5MTYwMDAwMDBaFw0y
OTAzMTgwMDAwMDBaMFsxCzAJBgNVBAYTAkJFMRkwFwYDVQQKExBHbG9iYWxTaWdu
IG52LXNhMTEwLwYDVQQDEyhHbG9iYWxTaWduIEdDQyBSMyBQZXJzb25hbFNpZ24g
MSBDQSAyMDIwMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvxvJBqEa
paux2/z3J7fFslROWjKVJ5rCMfWGsg17dmD7NSnG7Spoa8d3htXsls1IMxoO8Pyo
uQajNQqYmlYoxinlqenMNv7CJyEKMOAtglBmD6C/QC7kT+dSx4HfSTs8xmv8veJO
ldMzF8S/BEn/tD4w/Dvpg+oXOqDyOiHPTacRFK0QHoq5eEbBmVS8W0rwcaRotO9f
GTA+NjF0My7GLRNK0eMPGh2hcPZURQhXy7wRQ8XFIfEA6kaQHHN22ncnVtwqiTmA
wTR+4GNNVinG3KjNZLAVSnGrdCvT2I4Zo19hKy5PX6o7wrVXvMR4zV5VBFwV6ZDM
+xewao7Mup+SbwIDAQABo4IBiTCCAYUwDgYDVR0PAQH/BAQDAgGGMB0GA1UdJQQW
MBQGCCsGAQUFBwMCBggrBgEFBQcDBDASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1Ud
DgQWBBSFu/DMxDa1CmJ2o5kuj7s6aq3FUTAfBgNVHSMEGDAWgBSP8Et/qC5FJK5N
UPpjmove4t0bvDB6BggrBgEFBQcBAQRuMGwwLQYIKwYBBQUHMAGGIWh0dHA6Ly9v
Y3NwLmdsb2JhbHNpZ24uY29tL3Jvb3RyMzA7BggrBgEFBQcwAoYvaHR0cDovL3Nl
Y3VyZS5nbG9iYWxzaWduLmNvbS9jYWNlcnQvcm9vdC1yMy5jcnQwNgYDVR0fBC8w
LTAroCmgJ4YlaHR0cDovL2NybC5nbG9iYWxzaWduLmNvbS9yb290LXIzLmNybDBM
BgNVHSAERTBDMEEGCSsGAQQBoDIBKDA0MDIGCCsGAQUFBwIBFiZodHRwczovL3d3
dy5nbG9iYWxzaWduLmNvbS9yZXBvc2l0b3J5LzANBgkqhkiG9w0BAQsFAAOCAQEA
WWtqju12g524FdD2HwUXU1rSxeM5aSU1cUC1V/xBjXW0IjA7/3/vG2cietPPP/g3
lpoQePVJpQAKZml81fHwPPivFK9Ja41jJkgqGzkORSC0xYkh2gGeQg1JVaCzcrRz
JElRjT442m6FpbLHCebxIHLu0WBNjLZreB6MYMaqdPL6ItbXtD/BU4k517cEuUbc
zoBFZArajq7oUBWXuroln5AMnRwVNwgJN4Np0s4kkJ94KepzbFOLzcbnfUB0+xT4
foXmbM0GmmcPGOy0qvqEHJsBwDZXDxIk8oqCnnLngi7N94Sn4eTcmpZ9NH2dDN1O
TEPVXgRG5X1pBcNtMWG6MA==
-----END CERTIFICATE-----`
}
