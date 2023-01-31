package models

import (
	"encoding/json"
	"time"
)

type PaymentEvent struct {
	ID                        string          `json:"id"`
	GeneratedOn               time.Time       `json:"generatedOn"`
	Type                      EventType       `json:"type"`
	PaymentInstructionVersion int             `json:"paymentInstructionVersion"`
	PaymentProvider           PaymentProvider `json:"paymentProvider"`
	Reason                    Reason          `json:"reason"`
	PaymentInstructionID      string          `json:"paymentInstructionId"`
}

func (p PaymentEvent) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

type PaymentProvider struct {
	ID      string `json:"id"`
	Details struct {
		PaymentID string `json:"paymentId"`
	} `json:"details"`
}

type Reason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details"`
}

type EventType string
