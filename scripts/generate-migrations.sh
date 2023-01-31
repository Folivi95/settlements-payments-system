#!/bin/sh -e

go install github.com/kevinburke/go-bindata/...@latest

go-bindata -pkg migrations -ignore bindata -nometadata -prefix internal/adapters/payment_store/postgresql/migrations/ -o ./internal/adapters/payment_store/postgresql/migrations/bindata.go ./internal/adapters/payment_store/postgresql/migrations