#!/bin/bash

echo "build server ..." && \
go build -o cmd/gophermart/gophermart cmd/gophermart/main.go && \
echo "starting gophermarttest ..." && \
./gophermarttest \
            -test.v -test.run=^TestGophermart$ \
            -gophermart-binary-path=cmd/gophermart/gophermart \
            -gophermart-host=localhost \
            -gophermart-port=8080 \
            -gophermart-database-uri="postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${COMPOSE_PROJECT_NAME}_db/${POSTGRES_DB}?sslmode=disable" \
            -accrual-binary-path=cmd/accrual/accrual_linux_amd64 \
            -accrual-host=localhost \
            -accrual-port=8090 \
            -accrual-database-uri="postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${COMPOSE_PROJECT_NAME}_db/${POSTGRES_DB}?sslmode=disable" \

#            -gophermart-database-uri="postgresql://postgres:postgres@postgres/praktikum?sslmode=disable" \
#            -accrual-port=$(random unused-port) \
#            -accrual-database-uri="postgresql://postgres:postgres@postgres/praktikum?sslmode=disable"
