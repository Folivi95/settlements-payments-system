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

FROM deps as build
RUN /app/scripts/local-build.sh

FROM scratch
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/ ./
#COPY --from=build /app/*.env ./
#COPY --from=build /app/banking_circle_payment_service/ ./

CMD ["/app/settlements-payments-system"]
EXPOSE 8080
