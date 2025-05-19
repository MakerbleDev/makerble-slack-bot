FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=aarch64 go build -o my-go-app .

FROM alpine:latest

RUN apk add --no-cache tzdata && \
    cp /usr/share/zoneinfo/Asia/Kolkata /etc/localtime && \
    echo "Asia/Kolkata" > /etc/timezone

WORKDIR /root/

COPY --from=builder /app/my-go-app .

RUN chmod +x /root/my-go-app

EXPOSE 3000

CMD ["/root/my-go-app"]
