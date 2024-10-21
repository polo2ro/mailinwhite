#!/bin/sh

postmap /etc/postfix/sasl_passwd
chmod 600 /etc/postfix/sasl_passwd*

newaliases # generates a /etc/postfix/aliases.db

mkdir /var/spool/postfix/hold


addgroup app
adduser -S -h /opt/app -G app app
chown app:app /var/spool/postfix/hold

touch /var/log/mail.err
touch /var/log/mail.log

/usr/sbin/postfix -c /etc/postfix start
