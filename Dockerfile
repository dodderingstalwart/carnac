FROM golang:alpine3.22 as builder

WORKDIR /carnac

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o carnac .

FROM alpine:3.22

COPY --from=builder /carnac /usr/local/bin/carnac
CMD ["/carnac"]