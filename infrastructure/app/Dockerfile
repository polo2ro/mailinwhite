FROM golang:1.23.2 AS builder

COPY application/app /app
WORKDIR /app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./app ./src


FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/app .
ENTRYPOINT ["/app"]
