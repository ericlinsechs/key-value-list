# syntax=docker/dockerfile:1

## Build
FROM golang:1.19-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /my-app .

# # Make the executable file executable
# RUN chmod +x /main

## Deploy
FROM gcr.io/distroless/base-debian10

COPY --from=build /my-app /my-app

USER nonroot:nonroot

EXPOSE 8000

ENTRYPOINT ["/my-app"]