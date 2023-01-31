package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/gorilla/mux"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/sqs"
	"github.com/saltpay/settlements-payments-system/internal/adapters/aws/ufx_downloader"
)

type action string

const (
	purgeAction action = "purge"
)

func (a action) valid() bool {
	switch a {
	case purgeAction:
		return true
	default:
		return false
	}
}

type InternalHandler struct {
	allowPurge    bool
	queues        sqs.QueueClientMapping
	ufxDownloader ufx_downloader.UfxDownloader
}

func NewInternalHandler(queues sqs.QueueClientMapping, allowPurge bool, ufxDownloader ufx_downloader.UfxDownloader) *InternalHandler {
	return &InternalHandler{
		queues:        queues,
		allowPurge:    allowPurge,
		ufxDownloader: ufxDownloader,
	}
}

func (i *InternalHandler) GetDlqInformation(w http.ResponseWriter, r *http.Request) {
	var (
		ctx                 = r.Context()
		queueName           = mux.Vars(r)["name"]
		queueClient, exists = i.queues[sqs.QueueName(queueName)]
	)

	ctx = zapctx.WithFields(ctx, zap.String("queue_name", queueName))

	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	dlqInformation, err := queueClient.PeekAllMessages(ctx)
	if err != nil {
		zapctx.Debug(ctx, "fail to peek messages", zap.Error(err))
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}

	setJSON(w)
	_ = json.NewEncoder(w).Encode(dlqInformation)
}

func (i *InternalHandler) GetDlqUrls(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	for queueName := range i.queues {
		if !queueName.IsDlq() {
			continue
		}

		queueUrl := urlFromQueue(queueName, r)
		_, _ = fmt.Fprintf(w, `<a href="%v" target="_blank"> %v</a> <br>`, queueUrl, queueName)
	}
}

func (i *InternalHandler) PurgeQueue(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	queueName := sqs.QueueName(name)

	values := r.URL.Query()
	action := action(values.Get("action"))
	if !action.valid() {
		http.Error(w, "Action not supported", http.StatusNotFound)
		return
	}

	queueClient, exists := i.queues[queueName]
	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if !i.allowPurge && !queueName.IsDlq() {
		http.Error(w, "Queue purge is disable for non Dead Letter Queues", http.StatusBadRequest)
		return
	}

	err := queueClient.Purge(r.Context())
	if err != nil {
		http.Error(w, "Error when trying to purge", http.StatusBadRequest)
		return
	}
}

func (i *InternalHandler) GetQueueAttributes(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	queueName := sqs.QueueName(name)

	queueClient, exists := i.queues[queueName]
	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	attr, err := queueClient.Attributes(r.Context())
	if err != nil {
		http.Error(w, "Error when trying to purge", http.StatusBadRequest)
		return
	}

	setJSON(w)
	_ = json.NewEncoder(w).Encode(attr)
}

func (i *InternalHandler) DownloadUfxFile(w http.ResponseWriter, r *http.Request) {
	date, foundDate := mux.Vars(r)["date"]
	if !foundDate {
		date = time.Now().Format("2006-01-02")
	}

	filetype := mux.Vars(r)["filetype"]
	foundFiletype := i.ufxDownloader.ValidFileType(filetype)

	var (
		url string
		err error
	)
	if !foundDate && !foundFiletype {
		// filename passed not filetype
		url, err = i.ufxDownloader.GetPresignedURLWithFilename(r.Context(), filetype)
	} else {
		url, err = i.ufxDownloader.GetPresignedURL(r.Context(), date, filetype)
	}

	if err != nil {
		if uerr, ok := err.(ufx_downloader.Error); ok {
			switch uerr.Code() {
			case ufx_downloader.BucketNotFound:
				http.Error(w, "bucket not found", http.StatusInternalServerError)
				return
			case ufx_downloader.FileNotFound:
				http.Error(w, "file not found", http.StatusNotFound)
				return
			case ufx_downloader.NoFileWithPrefix:
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			case ufx_downloader.FileTypeNotFound:
				http.Error(w, "invalid filetype", http.StatusBadRequest)
				return
			case ufx_downloader.InvalidDate:
				http.Error(w, "invalid date", http.StatusBadRequest)
				return
			}
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	_, _ = fmt.Fprint(w, url)
}

func urlFromQueue(queueName sqs.QueueName, r *http.Request) string {
	u := url.URL{
		Scheme: getScheme(r.Host),
		Host:   r.Host,
		Path:   path.Join(r.URL.Path, string(queueName)),
	}
	return u.String()
}

func (i *InternalHandler) GetTeamMembers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	team := []string{"Anastasiia Dalakishvili", "David Fontes", "Yusuf Papurcu", "David Elek", "Einar Olafsson", "Vanessa Virgitti", "Catarina Bombaca", "Minhal Khan", "Dhiren Brahmbhatt", "Akash Kurdekar", "Rui Santos", "Thomas Mathew"}
	jsonBytes, err := json.Marshal(&team)
	if err != nil {
		zapctx.Debug(r.Context(), "error converting team list to JSON", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(jsonBytes)
	if err != nil {
		zapctx.Debug(r.Context(), "Error when writing team members", zap.Error(err))
	}
}
