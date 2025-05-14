FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o my-go-app .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/my-go-app .

CMD ["./my-go-app"]
