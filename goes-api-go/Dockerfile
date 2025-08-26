FROM golang:1.22 AS builder

WORKDIR /src
COPY go.mod go.sum ./

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/goes-api

FROM alpine:latest

# Install tzdata package
RUN apk add --no-cache tzdata

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/goes-api /app/goes-api

EXPOSE 3000

ENV GIN_MODE=release

ENTRYPOINT [ "/app/goes-api" ]