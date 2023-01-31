#!/bin/sh -e
go build -o local-aws-setup /app/scripts/setup_local_aws.go
/app/local-aws-setup
/app/settlements-payments-system