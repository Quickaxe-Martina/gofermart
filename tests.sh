#!/bin/bash

export PATH=$PATH:$(go env GOPATH)/bin
goimports -w .
pkill -9 "gophermart"
pkill -9 "accrual"
go build -o ./cmd/gophermart/gophermart ./cmd/gophermart/main.go

chmod +x ./cmd/gophermart/gophermart
chmod +x ./gophermarttest
rm -rf ./data.json
rm -rf ./cmd/gophermart/data.json

export DATABASE_DSN="postgresql://test:test@127.0.0.1:5433/test?sslmode=disable"
# export DEV_MODE="true"

./cmd/wipedb/wipedb "postgres://test:test@localhost:5433/test?sslmode=disable"

# ./gophermarttest \
# -test.v -test.run=TestGophermart/TestUserAuth \
# -gophermart-binary-path=cmd/gophermart/gophermart \
# -gophermart-database-uri="postgresql://test:test@127.0.0.1:5433/test?sslmode=disable" \
# -gophermart-host=127.0.0.1 \
# -gophermart-port=8080 \
# -accrual-binary-path=cmd/accrual/accrual_linux_amd64 \
# -accrual-database-uri="postgresql://test:test@127.0.0.1:5433/accrual?sslmode=disable" \
# -accrual-host=127.0.0.1 \
# -accrual-port=8081

# ./gophermarttest \
# -test.v -test.run=TestGophermart/TestUserOrders \
# -gophermart-binary-path=cmd/gophermart/gophermart \
# -gophermart-database-uri="postgresql://test:test@127.0.0.1:5433/test?sslmode=disable" \
# -gophermart-host=127.0.0.1 \
# -gophermart-port=8080 \
# -accrual-binary-path=cmd/accrual/accrual_linux_amd64 \
# -accrual-database-uri="postgresql://test:test@127.0.0.1:5433/accrual?sslmode=disable" \
# -accrual-host=127.0.0.1 \
# -accrual-port=8081

# ./gophermarttest \
# -test.v -test.run=TestGophermart/TestEndToEnd  \
# -gophermart-binary-path=cmd/gophermart/gophermart \
# -gophermart-database-uri="postgresql://test:test@127.0.0.1:5433/test?sslmode=disable" \
# -gophermart-host=127.0.0.1 \
# -gophermart-port=8080 \
# -accrual-binary-path=cmd/accrual/accrual_linux_amd64 \
# -accrual-database-uri="postgresql://test:test@127.0.0.1:5433/accrual?sslmode=disable" \
# -accrual-host=127.0.0.1 \
# -accrual-port=8081

./gophermarttest \
-test.v  \
-gophermart-binary-path=cmd/gophermart/gophermart \
-gophermart-database-uri="postgresql://test:test@127.0.0.1:5433/test?sslmode=disable" \
-gophermart-host=127.0.0.1 \
-gophermart-port=8080 \
-accrual-binary-path=cmd/accrual/accrual_linux_amd64 \
-accrual-database-uri="postgresql://test:test@127.0.0.1:5433/test?sslmode=disable" \
-accrual-host=127.0.0.1 \
-accrual-port=8081