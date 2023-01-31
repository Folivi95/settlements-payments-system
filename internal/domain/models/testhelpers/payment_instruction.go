package testhelpers

import (
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

// ValidPaymentInstruction returns a valid PaymentInstruction as well as its JSON representation.
func ValidPaymentInstruction() (models.PaymentInstruction, string, error) {
	validPaymentInstructionJSON := getValidPaymentInstructionsJSON()[0]

	paymentInstruction, err := models.NewPaymentInstructionFromJSON([]byte(validPaymentInstructionJSON))
	if err != nil {
		return models.PaymentInstruction{}, "", err
	}

	return paymentInstruction, validPaymentInstructionJSON, nil
}

func ValidISLPaymentInstruction() (models.PaymentInstruction, string, error) {
	validPaymentInstructionJSON := getValidPaymentInstructionsJSON()[1]

	paymentInstruction, err := models.NewPaymentInstructionFromJSON([]byte(validPaymentInstructionJSON))
	if err != nil {
		return models.PaymentInstruction{}, "", err
	}

	return paymentInstruction, validPaymentInstructionJSON, nil
}

func getValidPaymentInstructionsJSON() []string {
	return []string{
		`{
		  	"id": "339aec00-771c-467e-a8c0-9056c6d2580a",
			"version": 1,
			"status": "CREATED",
			"paymentProvider": "banking_circle",
			"incomingInstruction": {
				"merchant": {
					"contractNumber": "9000000",
					"name": "Ms Big Shot Merchant",
					"email": "testemail@testmerchant.is",
						"address": {
							"country": "GBR",
							"city": "TestCity",
							"addressLine1": "Testaddress 9",
							"addressLine2": "240 TestCity"
						},
					"account": {
						"accountNumber": "GB33BUKB20201555555555",
						"swift": "050026000000",
						"country": "GB",
						"swiftReferenceNumber": ""
					}
				},
				"metadata": {
					"source": "Way4",
					"filename": "testfilename",
					"fileType": "UFX"
				},
				"payment": {
					"sender": {
						"name": "RB",
						"accountNumber": "DK7389009999910509",
						"branchCode": ""
					},
					"amount": "10",
					"currency": { "isoCode": "EUR", "isoNumber": "978" },
					"executionDate": "2021-06-18T00:00:00Z"
				}
			},
			"events": [
				{
					"type": "DOMAIN.RECEIVED",
					"createdOn": "2021-08-18T00:00:00Z"
				}
			]
		}`,
		`{
		  	"id": "339aec00-771c-467e-a8c0-9056c6d245645",
		  	"version": 1, 
			"status": "CREATED",
			"paymentProvider": "islandsbanki",
			"incomingInstruction": {
				"merchant": {
					"contractNumber": "9000001",
					"name": "Ms Small Shot Merchant",
					"email": "test@email.pt",
					"address": {
						"country": "ISL",
						"city": "TestCity",
						"addressLine1": "Testaddress 99",
						"addressLine2": "241 TestCity"
					},
					"account": {
						"accountNumber": "IS33BUKB20201555555556",
						"swift": "050026000001",
						"country": "IS",
						"swiftReferenceNumber": ""
					}
				},
				"metadata": {
					"source": "Way4",
					"filename": "amazing_file_name",
					"fileType": "UFX"
				},
				"payment": {
					"sender": {
						"name": "RB",
						"accountNumber": "DK7389009999910509",
						"branchCode": ""
						},
					"amount": "3",
					"currency": { "isoCode": "ISK", "isoNumber": "352" },
					"executionDate": "2021-06-18T00:00:00Z"
				}
			},
			"events": [
							{
								"type": "DOMAIN.RECEIVED",
								"createdOn": "2021-08-18T00:00:00Z"
							}
						]
		}`,
	}
}
