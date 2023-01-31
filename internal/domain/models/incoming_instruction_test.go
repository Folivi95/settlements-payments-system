//go:build unit
// +build unit

package models_test

import (
	"sort"
	"testing"

	"github.com/saltpay/settlements-payments-system/internal/domain/models/testhelpers"

	"github.com/matryer/is"

	"github.com/saltpay/settlements-payments-system/internal/domain/models"
)

func TestIncomingInstruction_ByCurrency(t *testing.T) {
	t.Run("Order all incoming instructions by currency", func(t *testing.T) {
		is := is.New(t)

		invalidCurrency := models.CUP

		incomingInstructions := []models.IncomingInstruction{
			{
				Payment: models.Payment{
					Currency: models.Currency{
						IsoCode: models.GBP,
					},
				},
			},
			{
				Payment: models.Payment{
					Currency: models.Currency{
						IsoCode: invalidCurrency,
					},
				},
			},
			{
				Payment: models.Payment{
					Currency: models.Currency{
						IsoCode: models.CZK,
					},
				},
			},
			{
				Payment: models.Payment{
					Currency: models.Currency{
						IsoCode: models.RON,
					},
				},
			},
			{
				Payment: models.Payment{
					Currency: models.Currency{
						IsoCode: models.HUF,
					},
				},
			},
		}

		sort.Sort(models.ByCurrencyPriority(incomingInstructions))

		is.Equal(incomingInstructions[0].IsoCode(), models.CZK)
		is.Equal(incomingInstructions[1].IsoCode(), models.RON)
		is.Equal(incomingInstructions[2].IsoCode(), models.HUF)
		is.Equal(incomingInstructions[3].IsoCode(), models.GBP)
		is.Equal(incomingInstructions[4].IsoCode(), invalidCurrency)
	})
}

func TestIncomingInstruction_CurrencyTotals(t *testing.T) {
	t.Run("Return the total for the currency", func(t *testing.T) {
		is := is.New(t)

		incomingInstructions := models.IncomingInstructions{
			{
				Payment: models.Payment{
					Amount: "100.11",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
			{
				Payment: models.Payment{
					Amount: "200.22",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
		}

		summary, err := incomingInstructions.SumByCurrency()

		is.NoErr(err)
		is.Equal(summary[0], models.IncomingInstructionsSummary{
			CurrencyCode: models.EUR,
			Counter:      2,
			Amount:       300.33,
		})
	})

	t.Run("Returns sum of 2 types of currencies", func(t *testing.T) {
		is := is.New(t)

		incomingInstructions := models.IncomingInstructions{
			{
				Payment: models.Payment{
					Amount: "100",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
			{
				Payment: models.Payment{
					Amount: "200",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
			{
				Payment: models.Payment{
					Amount: "300",
					Currency: models.Currency{
						IsoCode: models.GBP,
					},
				},
			},
			{
				Payment: models.Payment{
					Amount: "400",
					Currency: models.Currency{
						IsoCode: models.GBP,
					},
				},
			},
		}

		summary, err := incomingInstructions.SumByCurrency()

		is.NoErr(err)
		is.Equal(len(summary), 2)
	})

	t.Run("Return total sum to 4 decimal places", func(t *testing.T) {
		is := is.New(t)

		incomingInstructions := models.IncomingInstructions{
			{
				Payment: models.Payment{
					Amount: "100.1111",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
			{
				Payment: models.Payment{
					Amount: "200.2222",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
		}

		summary, err := incomingInstructions.SumByCurrency()

		is.NoErr(err)
		is.Equal(summary[0], models.IncomingInstructionsSummary{
			CurrencyCode: models.EUR,
			Counter:      2,
			Amount:       300.3333,
		})
	})
}

func TestIncomingInstruction_FilterOutCurrency(t *testing.T) {
	t.Run("Filter out a single currency", func(t *testing.T) {
		is := is.New(t)

		incomingInstructions := models.IncomingInstructions{
			{
				Payment: models.Payment{
					Amount: "100.1",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
			{
				Payment: models.Payment{
					Amount: "200.2",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
			{
				Payment: models.Payment{
					Amount: "9900",
					Currency: models.Currency{
						IsoCode: models.ISK,
					},
				},
			},
			{
				Payment: models.Payment{
					Amount: "300",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
		}

		is.Equal(len(incomingInstructions), 4)
		incomingInstructions = incomingInstructions.FilterOutCurrency(models.EUR)
		is.Equal(len(incomingInstructions), 1)
	})
}

func TestIncomingInstruction_ReturnCurrency(t *testing.T) {
	t.Run("Returns all instructions for a currency", func(t *testing.T) {
		is := is.New(t)

		incomingInstructions := models.IncomingInstructions{
			{
				Payment: models.Payment{
					Amount: "100.1",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
			{
				Payment: models.Payment{
					Amount: "200.2",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
			{
				Payment: models.Payment{
					Amount: "9900",
					Currency: models.Currency{
						IsoCode: models.ISK,
					},
				},
			},
			{
				Payment: models.Payment{
					Amount: "300",
					Currency: models.Currency{
						IsoCode: models.EUR,
					},
				},
			},
		}

		is.Equal(len(incomingInstructions), 4)
		incomingInstructions = incomingInstructions.ReturnCurrency(models.EUR)
		is.Equal(len(incomingInstructions), 3)
	})
}

func TestIncomingInstruction_NormaliseAccountNumber(t *testing.T) {
	t.Run("returns a normalised account number", func(t *testing.T) {
		is := is.New(t)

		testCases := []struct {
			incomingAccountNumber   string
			normalisedAccountNumber string
			testDescription         string
		}{
			{"GB 33 BUKB 2020 1555 5555 55", "GB33BUKB20201555555555", "when it has spaces"},
			{"", "", "when iban is empty"},
			{"gb 33 BUkb 2020 1555 5555 55", "GB33BUKB20201555555555", "when it has a mix of upper/lowercase and spaces"},
			{"GB33BUKB20201555555555", "GB33BUKB20201555555555", "even when it's correct already"},
		}

		for _, testCase := range testCases {
			t.Run(testCase.testDescription, func(t *testing.T) {
				incomingInstruction := testhelpers.NewIncomingInstructionBuilder().WithMerchantAccountNumber(testCase.incomingAccountNumber).Build()
				newIncomingInstruction := incomingInstruction.NormaliseAccountNumber()
				is.Equal(newIncomingInstruction.AccountNumber(), testCase.normalisedAccountNumber)
			})
		}
	})
}
