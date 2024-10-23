FROM golang:1.23.2 AS builder

COPY application/postfix_hook /postfix_hook
WORKDIR /postfix_hook
RUN go mod download
RUN go build -o ./postfix_hook ./src


FROM alpine:3.20.3
COPY --from=builder /postfix_hook/postfix_hook /
RUN apk add --no-cache \
  libc6-compat \
  postfix \
  rsyslog \
  supervisor

COPY infrastructure/postfix/etc /etc
COPY infrastructure/postfix/entrypoint.sh /entrypoint.sh
RUN chmod +x entrypoint.sh

ENV SMTP_HOST=mailcatcher
ENV SMTP_PORT=1025
ENV SMTP_LOGIN=""
ENV SMTP_PASSWORD=""

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 25
HEALTHCHECK CMD postfix status
