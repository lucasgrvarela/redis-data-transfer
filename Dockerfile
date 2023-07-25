FROM golang:1.20

WORKDIR /app

COPY . .

RUN go build

ENTRYPOINT ["/app/redis-data-transfer"]
