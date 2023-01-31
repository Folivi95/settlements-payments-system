//go:build unit
// +build unit

package single_payment_endpoint_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/models/single_payment_endpoint"
)

func TestRequestDto_ToJSON(t *testing.T) {
	amounts := []float64{1346993.3, 0.01, 0.1, 0.99, 0.09, 0.9}

	for _, amount := range amounts {
		t.Run("With the given amount correctly serializes to JSON", func(t *testing.T) {
			stringAmount := strconv.FormatFloat(amount, 'f', -1, 64)
			expectedRequestJSON := fmt.Sprintf("{\"requestedExecutionDate\":\"0001-01-01T00:00:00Z\",\"debtorAccount\":{\"account\":\"\",\"financialInstitution\":\"\",\"country\":\"\"},\"debtorViban\":\"\",\"debtorReference\":\"\",\"debtorNarrativeToSelf\":\"\",\"currencyOfTransfer\":\"\",\"amount\":{\"currency\":\"\",\"amount\":%s},\"chargeBearer\":\"\",\"remittanceInformation\":{\"line1\":\"\",\"line2\":\"\",\"line3\":\"\",\"line4\":\"\"},\"creditorId\":\"\",\"creditorAccount\":{\"account\":\"\",\"financialInstitution\":\"\",\"country\":\"\"},\"creditorName\":\"\",\"creditorAddress\":{\"line1\":\"\",\"line2\":\"\",\"line3\":\"\"}}", stringAmount)

			requestDTO := single_payment_endpoint.RequestDto{Amount: single_payment_endpoint.Amount{Amount: amount}}
			requestJSON := requestDTO.ToJSON()

			assert.Equal(t, expectedRequestJSON, string(requestJSON))
		})
	}
}
