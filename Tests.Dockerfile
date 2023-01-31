FROM public.ecr.aws/docker/library/golang:1.17.8 as deps

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
