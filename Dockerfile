FROM golang:alpine AS builder

WORKDIR /app

## Copy master
COPY . .
ENV GOARCH=arm64 GOOS=linux CGO_ENABLED=0

## Check packages and build in arm based processor
RUN go mod tidy
RUN go build -v -o chisa cmd/main.go

## Use alpine image to execute binary
FROM alpine:latest

WORKDIR /app

## Copy the compiled Go binary from the builder stage
COPY --from=builder /app .
RUN chmod +x chisa

## Run Chisa
CMD ["./chisa"]
