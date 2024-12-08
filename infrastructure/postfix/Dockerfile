FROM golang:1.23.2 AS builder

COPY application/postfix_hook /postfix_hook
COPY application/libs /libs
WORKDIR /postfix_hook
RUN go mod download
RUN go build -o ./postfix_hook ./src


FROM alpine:3.20.3
RUN addgroup filter
RUN adduser -S -h /home/filter -G filter filter
COPY --from=builder /postfix_hook/postfix_hook /home/filter
COPY --from=builder /postfix_hook/templates /home/filter/templates
COPY infrastructure/postfix/home/filter/test_postfix_hook.sh /home/filter
RUN chown filter:filter /home/filter/postfix_hook
RUN chown filter:filter /home/filter/test_postfix_hook.sh && chmod +x /home/filter/test_postfix_hook.sh

RUN apk add --no-cache \
  libc6-compat \
  postfix \
  rsyslog \
  supervisor

COPY infrastructure/postfix/etc /etc
COPY infrastructure/postfix/entrypoint.sh /entrypoint.sh
RUN chmod +x entrypoint.sh

ENV MYHOSTNAME="host.example.com"

# to require TLS by the target SMTP server: "encrypt" 
# for mailcatcher, use "may"
ENV SECURITY_LEVEL="may"

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 25
HEALTHCHECK CMD postfix status
