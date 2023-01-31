package handlers

import (
	"net/http"

	"github.com/saltpay/settlements-payments-system/internal/adapters/replay_payment"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type ReplayPaymentHandler struct {
	replayPayment ports.ReplayPayment
}

func NewReplayPaymentHandler(replayPayment ports.ReplayPayment) *ReplayPaymentHandler {
	return &ReplayPaymentHandler{
		replayPayment: replayPayment,
	}
}

func (rp *ReplayPaymentHandler) ReplayMissingFundsPayments(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	action := values.Get("action")
	currency := values.Get("currency")
	file := values.Get("file")

	if isParameterEmpty(action) || isParameterEmpty(currency) || isParameterEmpty(file) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch action {
	case replay_payment.PayCurrencyFromFile:
		err := rp.replayPayment.Execute(r.Context(), currency, file)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func isParameterEmpty(parameter string) bool {
	return len(parameter) == 0
}
