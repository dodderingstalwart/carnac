FROM golang:1.24.5-alpine AS build

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN adduser -D -s /bin/sh appuser

WORKDIR /root/

COPY --from=build /app/main .

RUN chown appuser:appuser main

USER appuser

ENV DBUSER=root
ENV DBPASSWORD=password
ENV DBHOST=localhost
ENV DBPORT=3306
ENV DBNAME=carnac

CMD ["./main"]