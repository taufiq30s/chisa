FROM golang:1.21 as builder

WORKDIR /app

## Copy master and install that packages
COPY . .
RUN go.mod download

## Copy master then build in arm based processor
RUN env GOOS=linux GOARCH=arm64 go build -v -o chisa cmd/main.go

## Use alpine image to execute binary
FROM alpine:latest

WORKDIR /app

## Copy the compiled Go binary from the builder stage
COPY --from=builder /app/app .

## Run Chisa
CMD ["./chisa"]
