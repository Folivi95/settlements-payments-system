package single_payment_endpoint

import (
	"encoding/json"
	"time"
)

// RequestDto is the representation of the request body submitted to the
// single payment endpoint of the Banking Circle API.
type RequestDto struct {
	RequestedExecutiondate time.Time             `json:"requestedExecutionDate"`
	DebtorAccount          DebtorAccount         `json:"debtorAccount"`
	DebtorViban            string                `json:"debtorViban"`
	DebtorReference        string                `json:"debtorReference"`
	DebtorNarrativeToSelf  string                `json:"debtorNarrativeToSelf"`
	CurrencyOfTransfer     string                `json:"currencyOfTransfer"`
	Amount                 Amount                `json:"amount"`
	ChargeBearer           string                `json:"chargeBearer"`
	RemittanceInformation  RemittanceInformation `json:"remittanceInformation"`
	CreditorID             string                `json:"creditorId"`
	CreditorAccount        CreditorAccount       `json:"creditorAccount"`
	CreditorName           string                `json:"creditorName"`
	CreditorAddress        CreditorAddress       `json:"creditorAddress"`
}

func (r RequestDto) ToJSON() []byte {
	out, _ := json.Marshal(r)
	return out
}

type DebtorAccount struct {
	Account              string `json:"account"`
	FinancialInstitution string `json:"financialInstitution"`
	Country              string `json:"country"`
}

type Amount struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type RemittanceInformation struct {
	Line1 string `json:"line1"`
	Line2 string `json:"line2"`
	Line3 string `json:"line3"`
	Line4 string `json:"line4"`
}

type CreditorAccount struct {
	Account              string `json:"account"`
	FinancialInstitution string `json:"financialInstitution"`
	Country              string `json:"country"`
}

type CreditorAddress struct {
	Line1 string `json:"line1"`
	Line2 string `json:"line2"`
	Line3 string `json:"line3"`
}
