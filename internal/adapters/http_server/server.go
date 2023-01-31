package http_server

import (
	"net/http"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws"

	"github.com/saltpay/settlements-payments-system/internal/adapters/http_server/handlers/tests"

	"github.com/gorilla/mux"
	tracing "github.com/saltpay/go-mux-tracing"

	ports2 "github.com/saltpay/settlements-payments-system/banking_circle_payment_service/domain/ports"

	mainConfig "github.com/saltpay/settlements-payments-system/cmd/config"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_downloader"
	"github.com/saltpay/settlements-payments-system/internal/adapters/http_server/handlers"
	"github.com/saltpay/settlements-payments-system/internal/adapters/http_server/middleware"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

// todo: we can refactor the server to avoid our current regex middleware approach

func NewSettlementsPaymentsServer(
	config ServerConfig,
	r *mux.Router,
	makePayment ports.MakePayment,
	getPaymentInstruction ports.GetPaymentInstruction,
	getPaymentReport ports.GetPaymentReport,
	queues sqs.QueueClientMapping,
	getBCRejectionReport ports2.GetBankingCircleRejectionReport,
	accessControlMiddleware middleware.HTTPMiddleware,
	replayPayment ports.ReplayPayment,
	allowSqsPurge bool,
	ufxDownloader ufx_downloader.UfxDownloader,
	ufxUploader aws.UfxFileUploader,
) (server *http.Server) {
	paymentHandler := handlers.NewPaymentHandler(makePayment, getPaymentInstruction, getPaymentReport, getBCRejectionReport)
	replayPaymentHandler := handlers.NewReplayPaymentHandler(replayPayment)
	internalHandler := handlers.NewInternalHandler(queues, allowSqsPurge, ufxDownloader)
	testHandler := tests.NewHandler(ufxUploader)

	tracing.Enable(r, mainConfig.TKI())

	r.Handle("/payments", http.HandlerFunc(paymentHandler.PostPaymentInstructions)).Methods(http.MethodPost)
	r.Handle("/payments", http.HandlerFunc(paymentHandler.PostPaymentInstructions)).Queries("kafka", "{kafka}").Methods(http.MethodPost)
	r.Handle("/payments/report", http.HandlerFunc(paymentHandler.GetReport)).Methods(http.MethodGet)
	r.Handle("/payments/report/{date}", http.HandlerFunc(paymentHandler.GetReport)).Methods(http.MethodGet)
	r.Handle("/payments/currencies-report", http.HandlerFunc(paymentHandler.GetCurrencyReport)).Methods(http.MethodGet)
	r.Handle("/payments/currencies-report/{date}", http.HandlerFunc(paymentHandler.GetCurrencyReport)).Methods(http.MethodGet)
	r.Handle("/payments/{id}", http.HandlerFunc(paymentHandler.GetPaymentInstruction)).Methods(http.MethodGet)
	r.Handle("/payments/correlationId/{correlationId}", http.HandlerFunc(paymentHandler.GetPaymentInstructionByCorrelationID)).Methods(http.MethodGet)
	r.Handle("/mid/{mid}/{date}", http.HandlerFunc(paymentHandler.GetInstructionByMid)).Methods(http.MethodGet)

	r.Handle("/bc-report", http.HandlerFunc(paymentHandler.GetBCReport)).Methods(http.MethodGet)
	r.Handle("/bc-report/{date}", http.HandlerFunc(paymentHandler.GetBCReport)).Methods(http.MethodGet)

	r.Handle("/replay-payment", http.HandlerFunc(replayPaymentHandler.ReplayMissingFundsPayments)).Queries("action", "{action}", "currency", "{currency}", "file", "{file}").Methods(http.MethodPost)

	r.Handle("/internal/dead-letter-queues/{name}", http.HandlerFunc(internalHandler.GetDlqInformation)).Methods(http.MethodGet)
	r.Handle("/internal/dead-letter-queues", http.HandlerFunc(internalHandler.GetDlqUrls)).Methods(http.MethodGet)
	r.Handle("/internal/queues/{name}", http.HandlerFunc(internalHandler.PurgeQueue)).Queries("action", "{action}").Methods(http.MethodPost)
	r.Handle("/internal/queues/{name}/attributes", http.HandlerFunc(internalHandler.GetQueueAttributes)).Methods(http.MethodGet)
	r.Handle("/internal/ufx-file/{filetype}", http.HandlerFunc(internalHandler.DownloadUfxFile)).Methods(http.MethodGet)
	r.Handle("/internal/ufx-file/{filetype}/{date}", http.HandlerFunc(internalHandler.DownloadUfxFile)).Methods(http.MethodGet)
	r.Handle("/internal/team", http.HandlerFunc(internalHandler.GetTeamMembers)).Methods(http.MethodGet)

	r.Handle("/test/upload/file", http.HandlerFunc(testHandler.AddFileToS3Bucket)).Methods(http.MethodPost)

	server = &http.Server{
		Addr:         config.TCPAddress(),
		Handler:      accessControlMiddleware(r),
		ReadTimeout:  config.HTTPReadTimeout,
		WriteTimeout: config.HTTPWriteTimeout,
	}
	return
}
