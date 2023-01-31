package models

type CurrencyCode string

const (
	ISK CurrencyCode = "ISK"
	EUR CurrencyCode = "EUR"
)

var CurrenciesToIso = map[CurrencyCode]string{
	ISK: "352",
	EUR: "978",
}
