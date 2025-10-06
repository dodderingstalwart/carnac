FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o carnac .

FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/carnac .

RUN mkdir -p /app/datebase && chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

VOLUME ["/app/database"]

CMD ["./carnac"]