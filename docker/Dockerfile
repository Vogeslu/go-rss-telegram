# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.23 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN ls .

RUN CGO_ENABLED=0 GOOS=linux go build -o /application ./cmd/rss-telegram/main.go

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /application /application

USER nonroot:nonroot

ENTRYPOINT ["/application"]