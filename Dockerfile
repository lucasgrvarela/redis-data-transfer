FROM golang:1.20

WORKDIR /app

COPY . .

RUN go build

CMD ["/app/redis-data-transfer"]
