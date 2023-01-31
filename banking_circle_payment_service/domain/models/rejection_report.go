package models

import (
	"encoding/json"
	"time"
)

type RejectionReport struct {
	Rejections []Rejection `json:"rejections"`
}

type Rejection struct {
	PIdChannelUser         string    `json:"pIdChannelUser"`
	PTxndate               time.Time `json:"pTxndate"`
	ReportDate             time.Time `json:"reportDate"`
	CustomerID             string    `json:"customerId"`
	Account                string    `json:"account"`
	AccountCurrency        string    `json:"accountCurrency"`
	ValueDate              time.Time `json:"valueDate"`
	PaymentAmount          float64   `json:"paymentAmount"`
	PaymentCurrency        string    `json:"paymentCurrency"`
	TransferCurrency       string    `json:"transferCurrency"`
	DestinationIban        string    `json:"destinationIban"`
	PaymentReferenceNumber string    `json:"paymentReferenceNumber"`
	UserReferenceNumber    string    `json:"userReferenceNumber"`
	FileReferenceNumber    string    `json:"fileReferenceNumber"`
	SourceType             string    `json:"sourceType"`
	Status                 string    `json:"status"`
	StatusReason           string    `json:"statusReason"`
}

func (r RejectionReport) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

func NewRejectionReportFromJSON(in []byte) (RejectionReport, error) {
	var out RejectionReport
	err := json.Unmarshal(in, &out)
	return out, err
}
