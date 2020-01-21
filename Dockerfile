FROM golang:alpine AS builder

WORKDIR /build

COPY . .

RUN go build -v ./cmd/server/server.go


FROM postgres:12-alpine

WORKDIR /app

COPY --from=builder /build/server .
COPY conf/postgres.conf /etc/postgresql/12/main/postgresql.conf
COPY db/migrations/000_initial.sql /docker-entrypoint-initdb.d
COPY scripts/start.sh /docker-entrypoint-initdb.d

ENV POSTGRES_DSN=postgres://postgres:postgres@/postgres

EXPOSE 5000
