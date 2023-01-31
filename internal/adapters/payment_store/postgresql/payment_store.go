package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq"
	postgresTracing "github.com/saltpay/go-postgres-tracing"
	zapctx "github.com/saltpay/go-zap-ctx"
	"go.uber.org/zap"

	"github.com/saltpay/settlements-payments-system/internal/adapters/payment_store"
	"github.com/saltpay/settlements-payments-system/internal/domain/models"
	"github.com/saltpay/settlements-payments-system/internal/domain/ports"
)

type PostgresStore struct {
	db       *sql.DB
	observer payment_store.PaymentRepoObservability
}

const (
	storeInstructionQuery              = "storeInstruction"
	getInstructionQuery                = "getInstruction"
	getInstructionByCorrelationIDQuery = "getInstructionByCorrelationID"
	getReportQuery                     = "getReport"
	getCurrencyReportQuery             = "getCurrencyReport"
	getPaymentByMidQuery               = "getPaymentByMidQuery"
	updatePayment                      = "updatePayment"
)

var _ ports.GetPaymentInstructionFromRepo = PostgresStore{}

func NewPaymentStore(ctx context.Context, postgresConnString string, observer payment_store.PaymentRepoObservability) (PostgresStore, error) {
	driverName, err := postgresTracing.RegisterPostgresOTEL()
	if err != nil {
		return PostgresStore{}, fmt.Errorf("error parsing connection string: %w", err)
	}

	db, err := sql.Open(driverName, postgresConnString)
	if err != nil {
		return PostgresStore{}, fmt.Errorf("error opening db connection: %w", err)
	}

	if err := pingUntilAvailable(db); err != nil {
		return PostgresStore{}, err
	}

	pgxConfig, err := pgx.ParseConfig(postgresConnString)
	if err != nil {
		return PostgresStore{}, fmt.Errorf("error parsing connection string: %w", err)
	}

	migrationStart := time.Now()
	zapctx.Debug(ctx, "migration is started")
	if err := Execute(db, pgxConfig.Database); err != nil {
		return PostgresStore{}, fmt.Errorf("could not migrate schema: %w", err)
	}
	zapctx.Debug(ctx, "migration is done", zap.String("elapsedTime", time.Since(migrationStart).String()))

	return PostgresStore{db: db, observer: observer}, nil
}

func (s PostgresStore) CleanDBForTesting() error {
	_, err := s.db.Query(`delete from payment_instructions`)
	if err != nil {
		return fmt.Errorf("could not clean DB for testing, err: %v", err)
	}
	return nil
}

const (
	nPings      = 10
	backoffTime = 1 * time.Second
)

func pingUntilAvailable(db *sql.DB) error {
	var err error
	for i := 0; i < nPings; i++ {
		if err = db.Ping(); err == nil {
			return nil
		}
		time.Sleep(backoffTime)
	}
	return fmt.Errorf("failed to initialize db connection: %v", err)
}

// Ping checks if the sql connection is active.
func (s PostgresStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s PostgresStore) Store(ctx context.Context, instruction models.PaymentInstruction) error {
	ctx, span := postgresTracing.SpanWithContext(ctx, storeInstructionQuery)
	defer postgresTracing.EndSpan(span)

	s.observer.ReceivedStoreInstruction(ctx, instruction.ID())

	if len(instruction.IncomingInstruction.PaymentCorrelationId) == 0 {
		instruction.IncomingInstruction.PaymentCorrelationId = uuid.New().String()
	}
	startTime := time.Now()

	if instruction.GetStatus() != models.Failed && instruction.GetStatus() != models.Rejected {
		query := `SELECT body FROM payment_instructions 
            WHERE body->'incomingInstruction'->'merchant'->>'contractNumber' = $1 
            AND body->'incomingInstruction'->'merchant'->'account'->>'accountNumber'= $2 
            AND body->'incomingInstruction'->'payment'->>'amount' = $3 
            AND body->'incomingInstruction'->'payment'->'currency' ->> 'isoCode' =  $4 
			AND TO_DATE(body->'incomingInstruction'->'payment'->>'executionDate', 'YYYY-MM-DD') = TO_DATE($5, 'YYYY-MM-DD')
            AND body->>'status' = ANY($6) LIMIT 1`

		rows, err := s.db.Query(query,
			instruction.IncomingInstruction.Merchant.ContractNumber,
			instruction.IncomingInstruction.Merchant.Account.AccountNumber,
			instruction.IncomingInstruction.Payment.Amount,
			instruction.IncomingInstruction.Payment.Currency.IsoCode,
			instruction.IncomingInstruction.Payment.ExecutionDate,
			pq.Array([]models.PaymentInstructionStatus{models.SubmittedForProcessing, models.Successful, models.Received}),
		)
		if err != nil {
			return fmt.Errorf("unable to execute database query, err %w", err)
		}

		hasDuplication := rows.Next()
		if hasDuplication {
			return ErrDuplicate
		}

		err = rows.Close()
		if err != nil {
			return err
		}
	}

	paymentInstructionJSON, err := instruction.MarshalJSON()
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, "INSERT INTO payment_instructions(payment_instruction_id, body) VALUES($1, $2)", instruction.ID(), paymentInstructionJSON)
	if err != nil {
		err = fmt.Errorf("unable to insert payment instruction, err: %w", err)
		s.observer.FailedStore(ctx, instruction.ID(), instruction.ContractNumber(), err)
		return err
	}

	s.observer.StoreSuccessful(ctx, instruction.ID(), time.Since(startTime).Milliseconds())

	return nil
}

func (s PostgresStore) UpdatePayment(ctx context.Context, id models.PaymentInstructionID, status models.PaymentInstructionStatus, event models.PaymentInstructionEvent) error {
	ctx, span := postgresTracing.SpanWithContext(ctx, updatePayment)
	defer postgresTracing.EndSpan(span)

	updatedStatus := struct {
		Status models.PaymentInstructionStatus `json:"status"`
	}{Status: status}
	updatedStatusJson, err := json.Marshal(updatedStatus)
	if err != nil {
		return err
	}

	updatedEventJson, err := json.Marshal(event)
	if err != nil {
		return err
	}

	startTime := time.Now()
	query := `UPDATE payment_instructions 
						SET body = body::jsonb 
							|| jsonb_set(body::jsonb, array['events'], (body->'events')::jsonb || $1::jsonb)
				    	|| $2::jsonb
				    	|| CONCAT('{"version":', COALESCE(body->>'version','0')::int + 1, '}')::jsonb
							WHERE payment_instruction_id = $3`
	_, err = s.db.ExecContext(ctx, query, updatedEventJson, updatedStatusJson, id)
	if err != nil {
		err := fmt.Errorf("unable to update payment instruction, err: %v", err)
		s.observer.FailedUpdate(ctx, id, err)
		return err
	}
	s.observer.UpdateSuccessful(ctx, id, time.Since(startTime).Milliseconds())

	return nil
}

func (s PostgresStore) Get(ctx context.Context, id models.PaymentInstructionID) (models.PaymentInstruction, error) {
	ctx, span := postgresTracing.SpanWithContext(ctx, getInstructionQuery)
	defer postgresTracing.EndSpan(span)

	s.observer.ReceivedGetInstruction(ctx, id)
	startTime := time.Now()
	row := s.db.QueryRowContext(ctx, "SELECT body FROM payment_instructions WHERE payment_instruction_id = $1", id)
	var body []byte

	if err := row.Scan(&body); err != nil {
		if err == sql.ErrNoRows {
			s.observer.PaymentInstructionNotFound(ctx, id)
			return models.PaymentInstruction{}, PaymentInstructionMissingError{ID: id}
		} else {
			s.observer.FailedGet(ctx, id, err)
			return models.PaymentInstruction{}, err
		}
	}

	paymentInstructionFromJSON, err := models.NewPaymentInstructionFromJSON(body)
	if err != nil {
		s.observer.FailedGet(ctx, id, err)
		return models.PaymentInstruction{}, err
	}

	s.observer.GetSuccessful(ctx, id, time.Since(startTime).Milliseconds())
	return paymentInstructionFromJSON, nil
}

func (s PostgresStore) GetFromCorrelationID(ctx context.Context, correlationId string) ([]models.PaymentInstruction, error) {
	ctx, span := postgresTracing.SpanWithContext(ctx, getInstructionByCorrelationIDQuery)
	defer postgresTracing.EndSpan(span)

	row, err := s.db.QueryContext(ctx, `SELECT body FROM payment_instructions WHERE body->'incomingInstruction'->>'paymentCorrelationId' = $1;`, correlationId)
	if err != nil {
		return nil, err
	}

	result := make([]models.PaymentInstruction, 0)
	for row.Next() {
		var body []byte
		err = row.Scan(&body)
		if err != nil {
			return nil, err
		}

		paymentInstructionFromJSON, err := models.NewPaymentInstructionFromJSON(body)
		if err != nil {
			return nil, err
		}
		result = append(result, paymentInstructionFromJSON)
	}

	if len(result) == 0 {
		return nil, PaymentInstructionMissingError{CorrelationID: correlationId}
	}

	return result, nil
}

func (s PostgresStore) GetReport(ctx context.Context, date time.Time) (models.PaymentReport, error) {
	ctx, span := postgresTracing.SpanWithContext(ctx, getReportQuery)
	defer postgresTracing.EndSpan(span)

	startTime := time.Now()
	maxDate := date.AddDate(0, 0, 1)
	row, err := s.db.QueryContext(ctx, `
				select body->>'status' as status, count(1) as count from payment_instructions
				where body->'incomingInstruction'->'payment'->>'executionDate' > $1 and body->'incomingInstruction'->'payment'->>'executionDate' < $2
				group by body->>'status';`, date.String(), maxDate.String())
	if err != nil {
		return models.PaymentReport{}, err
	}

	stats := make(map[models.PaymentInstructionStatus]uint)
	for row.Next() {
		var statusString models.PaymentInstructionStatus
		var count uint

		err = row.Scan(&statusString, &count)
		if err != nil {
			continue
		}
		stats[statusString] = count
	}

	failedPayments, failureStats, err := s.getFailures(ctx, date, maxDate)
	report := newPaymentReport(stats, failedPayments, failureStats)
	if err != nil {
		return models.PaymentReport{}, err
	}

	s.observer.GotReport(ctx, time.Since(startTime).Milliseconds())

	return report, nil
}

func (s PostgresStore) getFailures(ctx context.Context, date, maxDate time.Time) ([]models.FailedInstruction, map[models.DomainFailureReasonCode]uint, error) {
	row, err := s.db.QueryContext(ctx,
		`select body->>'id' as id, body->'incomingInstruction'->'payment'->'currency'->>'isoCode' as currency, body->'incomingInstruction'->'merchant'->>'contractNumber' as mid, body->'events' as events
				from payment_instructions 
				where (body->'incomingInstruction'->'payment'->>'executionDate' > $1 and body->'incomingInstruction'->'payment'->>'executionDate' < $2) and (body->>'status'=$3 or body->>'status'=$4 );`,
		date.String(),
		maxDate.String(),
		models.Rejected,
		models.Failed,
	)
	if err != nil {
		return []models.FailedInstruction{}, map[models.DomainFailureReasonCode]uint{}, err
	}

	var failedPayments []models.FailedInstruction
	failedStats := make(map[models.DomainFailureReasonCode]uint)
	for row.Next() {
		var id models.PaymentInstructionID
		var currency models.CurrencyCode
		var mid string
		var rawEvents []byte
		var events []failureEventFromPG

		err = row.Scan(&id, &currency, &mid, &rawEvents)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(rawEvents, &events); err != nil {
			return []models.FailedInstruction{}, map[models.DomainFailureReasonCode]uint{}, err
		}

		var reason models.DomainFailureReasonCode
		if len(events) < 1 {
			reason = ""
		} else {
			mostRecentEvent := events[len(events)-1]
			reason = mostRecentEvent.GetCode()
		}

		if _, ok := failedStats[reason]; !ok {
			failedStats[reason] = 0
		}
		failedStats[reason] = failedStats[reason] + 1
		failedPayment := models.FailedInstruction{
			ID:       id,
			Currency: currency,
			Mid:      mid,
			Reason:   reason,
		}

		failedPayments = append(failedPayments, failedPayment)
	}

	return failedPayments, failedStats, nil
}

func (s PostgresStore) GetCurrencyReport(ctx context.Context, date time.Time) (models.PaymentCurrencyReport, error) {
	ctx, span := postgresTracing.SpanWithContext(ctx, getCurrencyReportQuery)
	defer postgresTracing.EndSpan(span)

	maxDate := date.AddDate(0, 0, 1)
	row, err := s.db.QueryContext(ctx, `select
        body->'incomingInstruction'->'payment'->'currency'->>'isoCode' as currency
		,body->'incomingInstruction'->'merchant'->>'highRisk' as highRisk
        ,body->>'status' as status
        ,count(1) as count
		,sum(cast(body->'incomingInstruction'->'payment'->>'amount' as decimal)) as amount
		from payment_instructions
		where (body->'incomingInstruction'->'payment'->>'executionDate' > $1 and body->'incomingInstruction'->'payment'->>'executionDate' < $2)
		group by body->'incomingInstruction'->'payment'->'currency'->>'isoCode', body->'incomingInstruction'->'merchant'->>'highRisk', body->>'status'
		order by currency, status, count`, date.String(), maxDate.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return models.PaymentCurrencyReport{}, ReportMissingError{Date: date.String()}
		} else {
			return models.PaymentCurrencyReport{}, err
		}
	}

	report := make(map[string]models.CurrencyStats)
	for row.Next() {
		var isoCode string
		var statusString models.PaymentInstructionStatus
		var count uint
		var isHighRisk bool
		var amountFloat float64

		err = row.Scan(&isoCode, &isHighRisk, &statusString, &count, &amountFloat)
		if err != nil {
			continue
		}

		highRiskString := ""
		if isHighRisk {
			highRiskString = "_HR"
		}
		currencyStats, ok := report[isoCode+highRiskString]
		if !ok {
			currencyStats = models.CurrencyStats{}
		}
		if statusString == models.Successful {
			currencyStats.Successful = currencyStats.Successful + count
			currencyStats.SuccessfulAmount = currencyStats.SuccessfulAmount + amountFloat
		} else {
			currencyStats.Failures = currencyStats.Failures + count
			currencyStats.FailuresAmount = currencyStats.FailuresAmount + amountFloat
		}
		currencyStats.Total = currencyStats.Total + count
		currencyStats.TotalAmount = currencyStats.TotalAmount + amountFloat
		report[isoCode+highRiskString] = currencyStats
	}

	return models.PaymentCurrencyReport{CurrencyReport: report}, nil
}

func (s PostgresStore) GetPaymentByMid(ctx context.Context, mid string, date time.Time) (models.PaymentInstruction, error) {
	ctx, span := postgresTracing.SpanWithContext(ctx, getPaymentByMidQuery)
	defer postgresTracing.EndSpan(span)

	maxDate := date.AddDate(0, 0, 1)
	row := s.db.QueryRowContext(ctx,
		`select body from payment_instructions 
				where body->'incomingInstruction'->'merchant'->>'contractNumber' = $1 and (body->'incomingInstruction'->'payment'->>'executionDate' > $2 and body->'incomingInstruction'->'payment'->>'executionDate' < $3) `,
		mid,
		date.String(),
		maxDate.String(),
	)

	var body []byte

	if err := row.Scan(&body); err != nil {
		if err == sql.ErrNoRows {
			return models.PaymentInstruction{}, MidMissingError{Mid: mid, Date: date.Format("2006-01-02")}
		} else {
			return models.PaymentInstruction{}, err
		}
	}

	paymentInstructionFromJSON, err := models.NewPaymentInstructionFromJSON(body)
	if err != nil {
		return models.PaymentInstruction{}, err
	}

	return paymentInstructionFromJSON, nil
}

func newPaymentReport(stats map[models.PaymentInstructionStatus]uint, failedPayments []models.FailedInstruction, failureStats map[models.DomainFailureReasonCode]uint) models.PaymentReport {
	var report models.PaymentReport
	report.Stats = models.PaymentStats{
		Successful:             stats[models.Successful],
		Failed:                 stats[models.Failed],
		Rejected:               stats[models.Rejected],
		SubmittedForProcessing: stats[models.SubmittedForProcessing],
		Received:               stats[models.Received],
	}
	report.FailedPayments = models.FailedPayments{
		FailedStats: models.FailedStats{
			Rejected:         failureStats[models.RejectedPayment],
			StuckInPending:   failureStats[models.StuckPayment],
			Unhandled:        failureStats[models.Unhandled],
			TransportMishap:  failureStats[models.TransportMishap],
			NoSourceAcct:     failureStats[models.NoSourceAcct],
			FailedValidation: failureStats[models.FailedValidation],
			MissingFunds:     failureStats[models.MissingFunds],
		},
		FailedInstructions: failedPayments,
	}

	return report
}

// PaymentInstructionMissingError is returned when payment instruction referenced by PaymentInstructionID can't be found.
type PaymentInstructionMissingError struct {
	ID            models.PaymentInstructionID
	CorrelationID string
}

func (p PaymentInstructionMissingError) Error() string {
	id := string(p.ID)
	if id == "" {
		id = p.CorrelationID
	}

	return fmt.Sprintf("payment instruction %q is not found", id)
}

// ReportMissingError is returned when no data is found for the request date.
type ReportMissingError struct {
	Date string
}

func (r ReportMissingError) Error() string {
	return fmt.Sprintf("report date %s is not found", r.Date)
}

// MidMissingError is returned when no date is found for requested mid and date.
type MidMissingError struct {
	Mid  string
	Date string
}

func (m MidMissingError) Error() string {
	return fmt.Sprintf("Payment for contract number %s was not found for date %s", m.Mid, m.Date)
}
