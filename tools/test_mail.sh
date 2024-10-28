#!/bin/bash

set -e

# Configuration
from_address="sender4@example.com"
to_address="recipient@example.com"
smtp_server="127.0.0.1"
smtp_port="2025"

send_command() {
    echo "$1"
    sleep 1
}

sendmail() {
    BOUNDARY=$(date +%s|md5sum|awk '{print $1;}')
    
    (
        sleep 1
        send_command "EHLO localhost"
        send_command "MAIL FROM:<$from_address>"
        send_command "RCPT TO:<$to_address>"
        send_command "DATA"
        sleep 1
        cat <<EOF
From: $from_address
To: $to_address
Reply-To: $from_address
Subject: Test mail
Content-Type: multipart/mixed; boundary="$BOUNDARY"

This is a MIME formatted message.  If you see this text it means that your
email software does not support MIME formatted messages.

--$BOUNDARY
Content-Type: text/plain; charset=UTF-8; format=flowed
Content-Disposition: inline

Test mail

--$BOUNDARY

.
EOF
        sleep 1
        send_command "QUIT"
    ) | nc $smtp_server $smtp_port
}

sendmail
