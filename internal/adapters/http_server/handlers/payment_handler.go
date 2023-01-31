package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	ports2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"
	"github.com/saltpay/settlements-payments-system/internal/adapters/payment_store/postgresql"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type PaymentHandler struct {
	makePayment           ports.MakePayment
	getPaymentInstruction ports.GetPaymentInstruction
	getPaymentReport      ports.GetPaymentReport
	getBCRejectionReport  ports2.GetBankingCircleRejectionReport
}

type paymentResponse struct {
	ID string `json:"id"`
}

func NewPaymentHandler(
	makePayment ports.MakePayment,
	getPaymentInstruction ports.GetPaymentInstruction,
	getPaymentReport ports.GetPaymentReport,
	getBCRejectionReport ports2.GetBankingCircleRejectionReport,
) *PaymentHandler {
	return &PaymentHandler{
		makePayment:           makePayment,
		getPaymentInstruction: getPaymentInstruction,
		getPaymentReport:      getPaymentReport,
		getBCRejectionReport:  getBCRejectionReport,
	}
}

func (p *PaymentHandler) PostPaymentInstructions(w http.ResponseWriter, r *http.Request) {
	incomingInstruction, err := models.NewIncomingInstructionFromJSON(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := p.makePayment.Execute(r.Context(), incomingInstruction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(paymentResponse{ID: string(id)}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (p *PaymentHandler) GetPaymentInstruction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := mux.Vars(r)["id"]

	paymentInstruction, err := p.getPaymentInstruction.Execute(ctx, models.PaymentInstructionID(id))
	if err != nil {
		missingError, isMissingErr := err.(postgresql.PaymentInstructionMissingError)
		if isMissingErr {
			http.Error(w, missingError.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dto := models.NewPaymentInstructionDTO(paymentInstruction)
	setJSON(w)

	if err := json.NewEncoder(w).Encode(dto); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (p *PaymentHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	date, found := mux.Vars(r)["date"]
	if !found {
		date = time.Now().Format("2006-01-02")
	}

	dateFormatted, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get report: %v", err), http.StatusBadRequest)
		return
	}

	report, err := p.getPaymentReport.GetReport(r.Context(), p.roundedDateString(dateFormatted))
	if err != nil {
		missingError, isMissingErr := err.(postgresql.ReportMissingError)
		if isMissingErr {
			http.Error(w, missingError.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to get report: %v", err), http.StatusInternalServerError)
		return
	}
	setJSON(w)
	_ = json.NewEncoder(w).Encode(report)
}

func getScheme(host string) string {
	if strings.Contains(host, "localhost") {
		return "http"
	}
	return "https"
}

func (p *PaymentHandler) GetBCReport(w http.ResponseWriter, r *http.Request) {
	date, found := mux.Vars(r)["date"]
	if !found {
		date = time.Now().Format("2006-01-02")
	}
	report, err := p.getBCRejectionReport.Execute(date)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get report: %v", err), http.StatusInternalServerError)
		return
	}
	setJSON(w)
	_ = json.NewEncoder(w).Encode(report)
}

func (p *PaymentHandler) GetCurrencyReport(w http.ResponseWriter, r *http.Request) {
	date, found := mux.Vars(r)["date"]
	if !found {
		date = time.Now().Format("2006-01-02")
	}
	dateFormatted, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get report: %v", err), http.StatusBadRequest)
		return
	}
	report, err := p.getPaymentReport.GetCurrencyReport(r.Context(), p.roundedDateString(dateFormatted))
	if err != nil {
		missingError, isMissingErr := err.(postgresql.ReportMissingError)
		if isMissingErr {
			http.Error(w, missingError.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to get report: %v", err), http.StatusInternalServerError)
		return
	}
	setJSON(w)
	_ = json.NewEncoder(w).Encode(report)
}

func (p *PaymentHandler) roundedDateString(toRound time.Time) time.Time {
	return time.Date(toRound.Year(), toRound.Month(), toRound.Day(), 0, 0, 0, 0, toRound.Location())
}

func (p *PaymentHandler) GetInstructionByMid(w http.ResponseWriter, r *http.Request) {
	date, foundDate := mux.Vars(r)["date"]
	if !foundDate {
		date = time.Now().Format("2006-01-02")
	}
	dateFormatted, err := time.Parse("2006-01-02", date)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get payment instruction: %v", err), http.StatusBadRequest)
		return
	}

	mid, foundMid := mux.Vars(r)["mid"]
	if !foundMid {
		http.Error(w, fmt.Sprintf("failed to get payment instruction: %v", err), http.StatusBadRequest)
		return
	}

	paymentInstruction, err := p.getPaymentReport.GetPaymentByMid(r.Context(), mid, p.roundedDateString(dateFormatted))
	if err != nil {
		missingError, isMissingErr := err.(postgresql.MidMissingError)
		if isMissingErr {
			http.Error(w, missingError.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to get payment instruction: %v", err), http.StatusInternalServerError)
		return
	}
	setJSON(w)
	_ = json.NewEncoder(w).Encode(paymentInstruction)
}

func (p *PaymentHandler) GetPaymentInstructionByCorrelationID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	correlationID := mux.Vars(r)["correlationId"]

	paymentInstructions, err := p.getPaymentInstruction.RetrieveByCorrelationID(ctx, correlationID)
	if err != nil {
		missingError, isMissingErr := err.(postgresql.PaymentInstructionMissingError)
		if isMissingErr {
			http.Error(w, missingError.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := make([]models.PaymentInstructionDTO, 0, len(paymentInstructions))
	for _, paymentInst := range paymentInstructions {
		result = append(result, models.NewPaymentInstructionDTO(paymentInst))
	}

	setJSON(w)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func setJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
