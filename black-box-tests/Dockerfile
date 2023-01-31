FROM public.ecr.aws/docker/library/golang:1.17.0 as deps

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPRIVATE=github.com/saltpay/*

RUN git config --global url."git@github.com".insteadOf "https://github.com"

WORKDIR /app
COPY vendor ./vendor
COPY go.mod go.sum ./

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY banking_circle_payment_service/ ./banking_circle_payment_service/
COPY scripts/ ./scripts/
COPY black-box-tests/ ./black-box-tests/
COPY *.env ./

COPY SaltPay_Root_CA_01.pem /etc/ssl/certs/SaltPay_Root_CA_01.pem
RUN update-ca-certificates

# ensure source code is compiling
# -c compiles test files but it does not executes such tests
RUN go test -c github.com/saltpay/settlements-payments-system/black-box-tests --tags=blackbox_success

FROM deps
ENTRYPOINT ["go", "test", "-v", "-count=1", "-timeout=20m", "--tags=blackbox_success", "./black-box-tests"]
