package models

type SourceAccount struct {
	Currency       string         `json:"currency"`
	IsHighRisk     bool           `json:"isHighRisk"`
	AccountDetails AccountDetails `json:"accountDetails"`
}

type AccountDetails struct {
	Iban            string  `json:"iban"`
	AccountID       string  `json:"accountId"`
	MaxIntraDayLoan float64 `json:"max_intra_day_loan"`
}

type SourceAccounts []SourceAccount

// TODO: should currency have its own type here?
func (s SourceAccounts) FindAccountNumber(currency string, isHighRisk bool) (AccountDetails, bool) {
	for _, bcAccount := range s {
		if bcAccount.Currency == currency && bcAccount.IsHighRisk == isHighRisk {
			return bcAccount.AccountDetails, true
		}
	}

	return AccountDetails{}, false
}
