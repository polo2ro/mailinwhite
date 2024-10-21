FROM golang:1.23.2 AS builder

COPY application/app /app
WORKDIR /app
RUN go mod download
RUN go build -o ./app ./src


FROM alpine:3.20.3
WORKDIR /root/
COPY --from=builder /app/app .
EXPOSE 8081

CMD ["./app"]
