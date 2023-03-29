# syntax=docker/dockerfile:1

## Build
FROM golang:1.19-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app .

## Deploy
FROM gcr.io/distroless/base-debian10

COPY --from=build /app /app

USER nonroot:nonroot

ENTRYPOINT ["/app"]