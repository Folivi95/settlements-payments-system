package models

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

// TODO: remove this json tags from here ?

type IncomingInstruction struct {
	Merchant             Merchant `json:"merchant"`
	Metadata             Metadata `json:"metadata"`
	Payment              Payment  `json:"payment"`
	PaymentCorrelationId string   `json:"paymentCorrelationId"`
}

type Address struct {
	Country      string `json:"country"`
	City         string `json:"city"`
	AddressLine1 string `json:"addressLine1"`
	AddressLine2 string `json:"addressLine2"`
}

type Account struct {
	AccountNumber        string `json:"accountNumber"`
	Swift                string `json:"swift"`
	Country              string `json:"country"`
	SwiftReferenceNumber string `json:"swiftReferenceNumber"`
	BankCountry          string `json:"bankCountry"`
}

type Merchant struct {
	ContractNumber string  `json:"contractNumber"`
	RegNumber      string  `json:"regNumber"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	Address        Address `json:"address"`
	Account        Account `json:"account"`
	HighRisk       bool    `json:"highRisk"`
}

type Metadata struct {
	Source   string `json:"source"`
	Filename string `json:"filename"`
	FileType string `json:"fileType"`
	Sender   string `json:"sender"`
}

type Currency struct {
	IsoCode   CurrencyCode `json:"isoCode"`
	IsoNumber string       `json:"isoNumber"`
}

type Payment struct {
	Sender        Sender    `json:"sender"`
	Amount        string    `json:"amount"`
	Currency      Currency  `json:"currency"`
	ExecutionDate time.Time `json:"executionDate"`
}

type Sender struct {
	Name          string `json:"name"`
	AccountNumber string `json:"accountNumber"`
	BranchCode    string `json:"branchCode"`
}

func NewIncomingInstructionFromJSON(in io.Reader) (IncomingInstruction, error) {
	var incomingInstruction IncomingInstruction
	err := json.NewDecoder(in).Decode(&incomingInstruction)
	return incomingInstruction, err
}

func NewIncomingInstructionFromBytes(in []byte) (IncomingInstruction, error) {
	var incomingInstruction IncomingInstruction
	err := json.Unmarshal(in, &incomingInstruction)
	return incomingInstruction, err
}

func (i IncomingInstruction) ToJSON() ([]byte, error) {
	return json.Marshal(i)
}

func (i IncomingInstruction) IsoCode() CurrencyCode {
	return i.Payment.Currency.IsoCode
}

func (i IncomingInstruction) AccountNumber() string {
	return i.Merchant.Account.AccountNumber
}

type ByCurrencyPriority []IncomingInstruction

func (a ByCurrencyPriority) Len() int { return len(a) }
func (a ByCurrencyPriority) Less(i, j int) bool {
	return getPriorityByCurrency(a[i].IsoCode()) < getPriorityByCurrency(a[j].IsoCode())
}
func (a ByCurrencyPriority) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

var CurrencyPriorityList = map[CurrencyCode]int{
	PLN: 1,
	CZK: 2,
	RON: 3,
	NOK: 4,
	DKK: 5,
	HRK: 6,
	SEK: 7,
	HUF: 8,
	CHF: 9,
	EUR: 10,
	AUD: 11,
	CAD: 12,
	GBP: 13,
	USD: 14,
}

func getPriorityByCurrency(currency CurrencyCode) int {
	priorityNumber := CurrencyPriorityList[currency]
	// If currency is not mapped in priority map put it last
	if priorityNumber == 0 {
		return 200
	}
	return priorityNumber
}

type IncomingInstructions []IncomingInstruction

type IncomingInstructionsSummary struct {
	CurrencyCode CurrencyCode
	Counter      int
	Amount       float64
	HighRisk     bool
}

func (a IncomingInstructions) SumByCurrency() ([]IncomingInstructionsSummary, error) {
	type currencySummary struct {
		amount float64
		count  int
	}
	summary := make(map[CurrencyCode]*currencySummary)
	highRisk := a[0].Merchant.HighRisk

	for _, instruction := range a {
		// check if payments in this instruction set are high risk or not to ensure that all are of the same type.
		if highRisk != instruction.Merchant.HighRisk {
			return []IncomingInstructionsSummary{}, fmt.Errorf("incoming instructions should not contain a mix of HR and non HR payments")
		}

		amount, err := strconv.ParseFloat(instruction.Payment.Amount, 64)
		if err != nil {
			return []IncomingInstructionsSummary{}, err
		}
		if summary[instruction.IsoCode()] == nil {
			summary[instruction.IsoCode()] = &currencySummary{}
		}
		summary[instruction.IsoCode()].amount += amount
		summary[instruction.IsoCode()].count++
	}

	var response []IncomingInstructionsSummary
	for key, value := range summary {
		response = append(response, IncomingInstructionsSummary{
			CurrencyCode: key,
			Amount:       math.Ceil(value.amount*10000) / 10000,
			Counter:      value.count,
			HighRisk:     highRisk,
		})
	}

	return response, nil
}

func (a IncomingInstructions) FilterOutCurrency(currency CurrencyCode) IncomingInstructions {
	filteredSlice := IncomingInstructions{}
	for _, instruction := range a {
		if instruction.IsoCode() != currency {
			filteredSlice = append(filteredSlice, instruction)
		}
	}

	return filteredSlice
}

func (a IncomingInstructions) ReturnCurrency(currency CurrencyCode) IncomingInstructions {
	filteredSlice := IncomingInstructions{}
	for _, instruction := range a {
		if instruction.IsoCode() == currency {
			filteredSlice = append(filteredSlice, instruction)
		}
	}

	return filteredSlice
}

func (i IncomingInstruction) NormaliseAccountNumber() IncomingInstruction {
	accountNumber := i.AccountNumber()
	accountNumberWithoutWhiteSpace := strings.ReplaceAll(accountNumber, " ", "")
	newAccountNumber := strings.ToUpper(accountNumberWithoutWhiteSpace)

	i.Merchant.Account.AccountNumber = newAccountNumber
	return i
}
