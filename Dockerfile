FROM golang:1.24 AS builder

WORKDIR /app

COPY . .

RUN make build

FROM registry.access.redhat.com/ubi9:latest

COPY --from=builder /app/bin/itpe-report /usr/bin/itpe-report

ENTRYPOINT ["/usr/bin/itpe-report"]
